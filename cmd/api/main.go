package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/spendalt/backend/config"
	"github.com/spendalt/backend/internal/analytics"
	"github.com/spendalt/backend/internal/auth"
	"github.com/spendalt/backend/internal/budget"
	"github.com/spendalt/backend/internal/category"
	"github.com/spendalt/backend/internal/common"
	"github.com/spendalt/backend/internal/savings"
	"github.com/spendalt/backend/internal/transaction"
	"github.com/spendalt/backend/internal/user"
)

func main() {
	cfg := config.Load()

	db, err := common.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Repositories
	userRepo := user.NewRepository(db)
	txRepo := transaction.NewRepository(db)
	catRepo := category.NewRepository(db)
	budgetRepo := budget.NewRepository(db)
	savingsRepo := savings.NewRepository(db)
	analyticsRepo := analytics.NewRepository(db)
	refreshRepo := auth.NewRefreshTokenRepository(db)

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
	userService := user.NewService(userRepo)
	txService := transaction.NewService(txRepo, categorizer, txWorker)
	catService := category.NewService(catRepo)
	budgetService := budget.NewService(budgetRepo)
	savingsService := savings.NewService(savingsRepo)
	analyticsService := analytics.NewService(analyticsRepo)

	// Handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)
	txHandler := transaction.NewHandler(txService)
	catHandler := category.NewHandler(catService)
	budgetHandler := budget.NewHandler(budgetService)
	savingsHandler := savings.NewHandler(savingsService)
	analyticsHandler := analytics.NewHandler(analyticsService)

	app := fiber.New(fiber.Config{
		BodyLimit: 1 * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin,Content-Type,Authorization,X-Device",
	}))
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	setupRoutes(app, authHandler, userHandler, txHandler, catHandler, budgetHandler, savingsHandler, analyticsHandler, cfg.JWTSecret, tokenStore)

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
	jwtSecret string,
	tokenStore auth.TokenStore,
) {
	api := app.Group("/api/v1")

	authLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * 60 * 1000000000,
	})
	api.Post("/auth/signup", authLimiter, authHandler.Signup)
	api.Post("/auth/login", authLimiter, authHandler.Login)
	api.Post("/auth/refresh", authLimiter, authHandler.Refresh)

	protected := api.Group("", auth.AuthRequired(jwtSecret, tokenStore))

	protected.Post("/auth/logout", authHandler.Logout)

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

	protected.Post("/transactions/ingest/sms", txHandler.IngestSMS)
	protected.Post("/transactions/ingest/manual", txHandler.IngestManual)
	protected.Get("/transactions", txHandler.GetTransactions)

	protected.Get("/categories", catHandler.GetCategories)
	protected.Get("/categories/breakdown", catHandler.GetCategoryBreakdown)

	protected.Post("/budgets", budgetHandler.CreateBudget)
	protected.Get("/budgets", budgetHandler.GetBudgets)
	protected.Put("/budgets/:id", budgetHandler.UpdateBudget)
	protected.Delete("/budgets/:id", budgetHandler.DeleteBudget)

	protected.Post("/savings", savingsHandler.CreateGoal)
	protected.Get("/savings", savingsHandler.GetGoals)
	protected.Put("/savings/:id/progress", savingsHandler.UpdateProgress)
	protected.Delete("/savings/:id", savingsHandler.DeleteGoal)

	protected.Get("/analytics/insights", analyticsHandler.GetInsights)
	protected.Get("/analytics/weekly-trend", analyticsHandler.GetWeeklyTrend)

	protected.Get("/health/score", analyticsHandler.GetHealthScore)
}
