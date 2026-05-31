package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/moninte/backend/config"
	"github.com/moninte/backend/internal/analytics"
	"github.com/moninte/backend/internal/auth"
	"github.com/moninte/backend/internal/budget"
	"github.com/moninte/backend/internal/category"
	"github.com/moninte/backend/internal/common"
	"github.com/moninte/backend/internal/core"
	"github.com/moninte/backend/internal/dashboard"
	"github.com/moninte/backend/internal/lang"
	"github.com/moninte/backend/internal/migrations"
	"github.com/moninte/backend/internal/monitor"
	"github.com/moninte/backend/internal/payment"
	"github.com/moninte/backend/internal/savings"
	"github.com/moninte/backend/internal/subscription"
	"github.com/moninte/backend/internal/transaction"
	"github.com/moninte/backend/internal/user"
)

func main() {
	cfg := config.Load()

	// Explicit migration order — dependencies before dependents
	migrations.Ordered = []migrations.Domain{
		{Name: "user", FS: user.MigrationFiles},
		{Name: "auth", FS: auth.MigrationFiles},
		{Name: "monitor", FS: monitor.MigrationFiles},
		{Name: "transaction", FS: transaction.MigrationFiles},
		{Name: "category", FS: category.MigrationFiles},
		{Name: "budget", FS: budget.MigrationFiles},
		{Name: "savings", FS: savings.MigrationFiles},
		{Name: "subscription", FS: subscription.MigrationFiles},
	}

	db, err := common.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db.DB); err != nil {
		log.Fatal("Migrations failed:", err)
	}

	// Repositories
	userRepo := user.NewRepository(db)
	txRepo := transaction.NewRepository(db)
	catRepo := category.NewRepository(db)
	budgetRepo := budget.NewRepository(db)
	savingsRepo := savings.NewRepository(db)
	analyticsRepo := analytics.NewRepository(db)
	refreshRepo := auth.NewRefreshTokenRepository(db)
	subRepo := subscription.NewRepository(db)

	// Seed static data
	if err := catRepo.Seed(); err != nil {
		log.Printf("[warn] category seed failed: %v", err)
	}

	// Token store (Redis optional — falls back to in-memory)
	tokenStore, err := auth.NewTokenStore(cfg.RedisURL)
	if err != nil {
		log.Println("[warn] Token store init failed, using in-memory:", err)
		tokenStore, _ = auth.NewTokenStore("")
	}

	// Ingestion worker
	categorizer := transaction.NewRuleEngine()
	txWorker := transaction.NewWorker(txRepo, categorizer)
	ctx, cancel := context.WithCancel(context.Background())
	go txWorker.Run(ctx)

	// Services
	authService := auth.NewService(userRepo, cfg.JWTSecret, tokenStore, refreshRepo)
	userService := user.NewService(userRepo, refreshRepo)
	txService := transaction.NewService(txRepo, categorizer, txWorker)
	catService := category.NewService(catRepo)
	budgetService := budget.NewService(budgetRepo)
	savingsService := savings.NewService(savingsRepo)
	analyticsService := analytics.NewService(analyticsRepo)
	dashboardService := dashboard.NewService(db)
	subService := subscription.NewService(subRepo)

	paymentRegistry := payment.NewRegistry()
	if cfg.PaystackSecret != "" {
		paymentRegistry.Register(payment.NewPaystackProvider(cfg.PaystackSecret))
	}

	// Handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)
	txHandler := transaction.NewHandler(txService)
	catHandler := category.NewHandler(catService)
	budgetHandler := budget.NewHandler(budgetService)
	savingsHandler := savings.NewHandler(savingsService)
	analyticsHandler := analytics.NewHandler(analyticsService)
	dashboardHandler := dashboard.NewHandler(dashboardService)
	subHandler := subscription.NewHandler(subService, paymentRegistry)

	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB for statement uploads
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "internal server error"
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
				msg = e.Message
			}
			return c.Status(code).JSON(core.ErrorResponse{Error: msg})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin,Content-Type,Authorization,X-Device,X-App-Version",
	}))
	app.Use(monitor.SecurityHeaders())
	appLogger := monitor.NewLogger()
	monitorRepo := monitor.NewRepository(db, ctx)
	monitorHandler := monitor.NewHandler(monitorRepo)
	app.Use(monitor.RequestMonitor(appLogger, monitorRepo))

	app.Static("/uploads", "./uploads")

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	setupRoutes(app, authHandler, userHandler, txHandler, catHandler, budgetHandler, savingsHandler, analyticsHandler, monitorHandler, dashboardHandler, subHandler, subService, cfg.JWTSecret, tokenStore)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down...")
		cancel()
		_ = app.Shutdown()
	}()

	log.Printf("Server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Println("Server stopped:", err)
	}
}

func setupRoutes(
	app *fiber.App,
	authHandler *auth.Handler,
	userHandler *user.Handler,
	txHandler *transaction.Handler,
	catHandler *category.Handler,
	budgetHandler *budget.Handler,
	savingsHandler *savings.Handler,
	analyticsHandler *analytics.Handler,
	monitorHandler *monitor.Handler,
	dashboardHandler *dashboard.Handler,
	subHandler *subscription.Handler,
	subService subscription.Service,
	jwtSecret string,
	tokenStore auth.TokenStore,
) {
	api := app.Group("/api/v1")

	authLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * 60 * 1000000000,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(core.ErrorResponse{Error: lang.ErrRateLimited})
		},
	})

	ingestLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * 60 * 1000000000,
		KeyGenerator: func(c *fiber.Ctx) string {
			if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
				return "ingest:" + uid
			}
			return "ingest:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(core.ErrorResponse{Error: lang.ErrRateLimited})
		},
	})
	api.Post("/auth/signup", authLimiter, authHandler.Signup)
	api.Post("/auth/login", authLimiter, authHandler.Login)
	api.Post("/auth/refresh", authLimiter, authHandler.Refresh)
	api.Post("/auth/logout", authHandler.Logout)
	api.Post("/auth/oidc", authLimiter, authHandler.OIDCLogin)

	protected := api.Group("", auth.AuthRequired(jwtSecret, tokenStore))

	protected.Post("/user/avatar", userHandler.UploadAvatar)
	protected.Get("/user/profile", userHandler.GetProfile)
	protected.Put("/user/profile", userHandler.UpdateProfile)
	protected.Post("/user/change-password", userHandler.ChangePassword)
	protected.Delete("/user/account", userHandler.DeleteAccount)
	protected.Get("/user/preferences", userHandler.GetPreferences)
	protected.Put("/user/preferences", userHandler.SavePreferences)
	protected.Get("/user/linked-accounts", userHandler.GetLinkedAccounts)
	protected.Delete("/user/linked-accounts/:id", userHandler.RemoveLinkedAccount)
	protected.Post("/user/linked-accounts/:id/sync", userHandler.SyncLinkedAccount)
	protected.Get("/user/sessions", userHandler.GetSessions)
	protected.Delete("/user/sessions", userHandler.RevokeAllSessions)
	protected.Delete("/user/sessions/:id", userHandler.RevokeSession)

	protected.Get("/user/activity", monitorHandler.GetActivity)

	protected.Post("/transactions/ingest/sms", ingestLimiter, txHandler.IngestSMS)
	protected.Post("/transactions/ingest/sms/batch", ingestLimiter, txHandler.IngestSMSBatch)
	protected.Post("/transactions/ingest/manual", ingestLimiter, txHandler.IngestManual)
	protected.Post("/transactions/ingest/upload", ingestLimiter, txHandler.IngestUpload)
	protected.Get("/transactions", txHandler.GetTransactions)

	protected.Get("/categories", catHandler.GetCategories)
	protected.Get("/categories/breakdown", catHandler.GetCategoryBreakdown)

	protected.Post("/budgets", budgetHandler.CreateBudget)
	protected.Get("/budgets", budgetHandler.GetBudgets)
	protected.Put("/budgets/:id", budgetHandler.UpdateBudget)
	protected.Delete("/budgets/:id", budgetHandler.DeleteBudget)

	protected.Post("/savings", savingsHandler.CreateGoal)
	protected.Get("/savings", savingsHandler.GetComposite)
	protected.Get("/savings/list", savingsHandler.GetGoals)
	protected.Put("/savings/:id/progress", savingsHandler.UpdateProgress)
	protected.Delete("/savings/:id", savingsHandler.DeleteGoal)

	protected.Get("/analytics/insights", analyticsHandler.GetInsights)
	protected.Get("/analytics/weekly-trend", analyticsHandler.GetWeeklyTrend)
	protected.Get("/analytics", analyticsHandler.GetComposite)

	protected.Get("/dashboard", dashboardHandler.GetDashboard)

	protected.Get("/health/score", analyticsHandler.GetHealthScore)

	// Subscription
	api.Get("/plans", subHandler.GetPlans)
	protected.Get("/subscription", subHandler.GetSubscription)
	protected.Delete("/subscription", subHandler.Cancel)

	// Payment webhooks — no auth, signature-verified by provider
	api.Post("/webhooks/:provider", subHandler.Webhook)

	// Premium-gated routes
	premium := protected.Group("", subscription.RequireEntitlement(subService, "mono_sync"))
	_ = premium // future: premium.Post("/sync/mono", ...)
}
