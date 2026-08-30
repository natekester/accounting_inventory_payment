package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	accDB "github.com/natekester/inventory-payment-integration/backend/internal/modules/accounting/db"
	accDomain "github.com/natekester/inventory-payment-integration/backend/internal/modules/accounting/domain"
	accHandler "github.com/natekester/inventory-payment-integration/backend/internal/modules/accounting/handler"
	accRepo "github.com/natekester/inventory-payment-integration/backend/internal/modules/accounting/repository"
	accStrategy "github.com/natekester/inventory-payment-integration/backend/internal/modules/accounting/strategy"

	invDB "github.com/natekester/inventory-payment-integration/backend/internal/modules/inventory/db"
	invDomain "github.com/natekester/inventory-payment-integration/backend/internal/modules/inventory/domain"
	invHandler "github.com/natekester/inventory-payment-integration/backend/internal/modules/inventory/handler"
	invRepo "github.com/natekester/inventory-payment-integration/backend/internal/modules/inventory/repository"

	payDB "github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/db"
	payDomain "github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/domain"
	payHandler "github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/handler"
	payRepo "github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/repository"
	payStrategy "github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/strategy"

	"github.com/natekester/inventory-payment-integration/backend/internal/shared/database"
	"github.com/natekester/inventory-payment-integration/backend/internal/shared/eventbus"
)

func main() {
	log.Println("Starting Modular Monolith Backend...")

	// 1. Initialize Single Shared Database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// 2. Initialize In-Memory EventBus
	bus := eventbus.NewEventBus()

	// 3. Execute Module Schema Migrations
	log.Println("Running module database migrations...")
	if err := invDB.Migrate(db); err != nil {
		log.Fatalf("Inventory schema migration failed: %v", err)
	}
	if err := payDB.Migrate(db); err != nil {
		log.Fatalf("Payment schema migration failed: %v", err)
	}
	if err := accDB.Migrate(db); err != nil {
		log.Fatalf("Accounting schema migration failed: %v", err)
	}
	log.Println("Database migrations completed successfully.")

	// 4. Instantiate Repositories & Domain Services (Strategy Contexts)
	// Inventory Module
	iRepo := invRepo.NewGORMRepository(db)
	iService := invDomain.NewService(iRepo)
	iHandler := invHandler.NewHTTPHandler(iService)

	// Payment Module
	pRepo := payRepo.NewGORMRepository(db)
	pService := payDomain.NewService(pRepo, bus)
	pService.RegisterStrategy(payStrategy.NewStripeStrategy("sk_test_mock_key"))
	pHandler := payHandler.NewHTTPHandler(pService)

	// Accounting Module
	aRepo := accRepo.NewGORMRepository(db)
	aService := accDomain.NewService(aRepo, bus)
	aService.RegisterStrategy(accStrategy.NewRilletStrategy("rillet_mock_key"))
	aHandler := accHandler.NewHTTPHandler(aService)

	// 5. Setup Gin Router & Middleware
	router := gin.Default()

	// CORS Middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Healthcheck
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "modular-monolith-backend"})
	})

	// Register Module API Routes under /api/v1
	v1 := router.Group("/api/v1")
	iHandler.RegisterRoutes(v1)
	pHandler.RegisterRoutes(v1)
	aHandler.RegisterRoutes(v1)

	// 6. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
