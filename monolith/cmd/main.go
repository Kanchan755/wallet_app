package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kanchan755/wallet_app/monolith/internal/config"
	"github.com/kanchan755/wallet_app/monolith/internal/database"
	ledgerRepository "github.com/kanchan755/wallet_app/monolith/internal/ledger/repository"
	ledgerService "github.com/kanchan755/wallet_app/monolith/internal/ledger/service"
	"github.com/kanchan755/wallet_app/monolith/internal/logger"
	"github.com/kanchan755/wallet_app/monolith/internal/middleware"
	txHandler "github.com/kanchan755/wallet_app/monolith/internal/transaction/handler"
	txRepository "github.com/kanchan755/wallet_app/monolith/internal/transaction/repository"
	txService "github.com/kanchan755/wallet_app/monolith/internal/transaction/service"
	userHandler "github.com/kanchan755/wallet_app/monolith/internal/user/handler"
	userRepository "github.com/kanchan755/wallet_app/monolith/internal/user/repository"
	userService "github.com/kanchan755/wallet_app/monolith/internal/user/service"
	walletHandler "github.com/kanchan755/wallet_app/monolith/internal/wallet/handler"
	walletRepository "github.com/kanchan755/wallet_app/monolith/internal/wallet/repository"
	walletService "github.com/kanchan755/wallet_app/monolith/internal/wallet/service"
)

func main() {
	// initialize the log
	logger.InitLogger()
	logger.Log.Info("Starting the monolith application...")
	// Load configuration
	cfg := config.Loadconfig()

	// Connect to the database
	db, err := database.ConnectWithRetry(cfg.DBDSN, 5)
	if err != nil {
		logger.Log.Error("Could not connect to the database: %v", err)
	}
	defer db.Close()
	logger.Log.Info("Application Succesfully initilaized ")

	// connect to redis
	rdb, err := database.ConnectedRedis(cfg.RedisAddr)
	if err != nil {
		logger.Log.Error("ciritical Error: Could not connect to Redis", "error", err.Error())
	}
	defer rdb.Close()

	//1.initiate layer
	uRepo := userRepository.NewMySQLUserRepository(db)
	wRepo := walletRepository.NewMySQLWalletRepository(db)
	tRepo := txRepository.NewMySQLTransactionRepository(db)
	lRepo := ledgerRepository.NewLedgerRepository(db)

	//inject db to user service for transaction
	uSvc := userService.NewUserService(db, uRepo, wRepo)
	uHandler := userHandler.NewUserHandler(uSvc)
	wSvc := walletService.NewWalletService(wRepo, rdb)
	wHandler := walletHandler.NewWalletHandler(wSvc)
	tSvc := txService.NewTransactionService(db, rdb, tRepo, uRepo, wRepo, lRepo)
	tHandler := txHandler.NewTransactionHandler(tSvc)
	lSvc := ledgerService.NewLedgerService(lRepo, wRepo)
	_ = lSvc // Keep ledger service for future integration

	//2. Start the HTTP server
	//r := gin.Default()
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.ErrorHandler())

	// Define routes and handlers
	v1 := r.Group("/api/v1")
	{
		// Public routes
		v1.POST("/users/register", uHandler.RegisterUser)
		v1.POST("/users/login", uHandler.Login)
		v1.POST("/users", uHandler.RegisterUser)
		v1.GET("/users/:id", uHandler.GetUserProfile)
		v1.PUT("/users/:id", uHandler.UpdateUserProfile)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/users/me", uHandler.GetCurrentUser)
			protected.POST("/users/me/avatar", uHandler.UpdateAvatar)
			protected.GET("/wallets/me", wHandler.GetWalletByUserID)
			protected.POST("/transfer", tHandler.Transfer)
			protected.GET("/transactions/history", tHandler.GetHistory)
		}
	}

	// Start the server
	logger.Log.Info("Server running on port 8080")
	if err := r.Run(":8080"); err != nil {
		logger.Log.Error("Failed to start the server: %v", err)
	}

}
