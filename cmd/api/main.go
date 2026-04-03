package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/spendalt/backend/config"
	"github.com/spendalt/backend/internal/common"
	"github.com/spendalt/backend/internal/core"
	"github.com/spendalt/backend/internal/user"
	"github.com/spendalt/backend/internal/auth"
	"github.com/spendalt/backend/internal/transaction"
	"github.com/spendalt/backend/internal/category"
	"github.com/spendalt/backend/internal/budget"
	"github.com/spendalt/backend/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := common.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := user.NewRepository(db)
	txRepo := transaction.NewRepository(db)
	catRepo := category.NewRepository(db)
	budgetRepo := budget.NewRepository(db)
	budgetCoreRepo := &core.Repository[budget.Budget]{
		DB:    db,
		Table: "budgets",
		Scan: func(b *budget.Budget) []interface{} {
			return []interface{}{&b.ID, &b.UserID, &b.Category, &b.Amount, &b.Period, &b.CreatedAt}
		},
	}

	// Initialize services
	authService := auth.NewService(userRepo, cfg.JWTSecret)
	userService := user.NewService(userRepo)
	txService := transaction.NewService(txRepo)
	catService := category.NewService(catRepo)
	budgetService := budget.NewService(budgetRepo, budgetCoreRepo)

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)
	txHandler := transaction.NewHandler(txService)
	catHandler := category.NewHandler(catService)
	budgetHandler := budget.NewHandler(budgetService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Middleware
	app.Use(cors.New())
	app.Use(logger.New())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Setup routes
	setupRoutes(app, authHandler, userHandler, txHandler, catHandler, budgetHandler, cfg.JWTSecret)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func setupRoutes(
	app *fiber.App,
	authHandler *auth.Handler,
	userHandler *user.Handler,
	txHandler *transaction.Handler,
	catHandler *category.Handler,
	budgetHandler *budget.Handler,
	jwtSecret string,
) {
	// API routes
	api := app.Group("/api/v1")

	// Public routes
	api.Post("/auth/signup", authHandler.Signup)
	api.Post("/auth/login", authHandler.Login)

	// Protected routes
	protected := api.Group("", middleware.AuthRequired(jwtSecret))
	
	// User routes
	protected.Get("/user/profile", userHandler.GetProfile)
	protected.Put("/user/profile", userHandler.UpdateProfile)
	protected.Post("/user/change-password", userHandler.ChangePassword)
	protected.Delete("/user/account", userHandler.DeleteAccount)
	
	// Transaction routes
	protected.Post("/transactions/ingest/sms", txHandler.IngestSMS)
	protected.Post("/transactions/ingest/manual", txHandler.IngestManual)
	protected.Get("/transactions", txHandler.GetTransactions)
	
	// Category routes
	protected.Get("/categories", catHandler.GetCategories)
	protected.Get("/categories/breakdown", catHandler.GetCategoryBreakdown)
	
	// Budget routes
	protected.Post("/budgets", budgetHandler.CreateBudget)
	protected.Get("/budgets", budgetHandler.GetBudgets)
	protected.Put("/budgets/:id", budgetHandler.UpdateBudget)
	protected.Delete("/budgets/:id", budgetHandler.DeleteBudget)
}
