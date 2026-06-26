package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kanchan755/wallet_app/monolith/internal/config"
	"github.com/kanchan755/wallet_app/monolith/internal/database"
	"github.com/kanchan755/wallet_app/monolith/internal/logger"
	userHandler "github.com/kanchan755/wallet_app/monolith/internal/user/handler"
	userRepository "github.com/kanchan755/wallet_app/monolith/internal/user/repository"
	userService "github.com/kanchan755/wallet_app/monolith/internal/user/service"
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

	//1.initiate layer
	uRepo := userRepository.NewMySQLUserRepository(db)
	uSvc := userService.NewUserService(uRepo)
	uHandler := userHandler.NewUserHandler(uSvc)

	//2. Start the HTTP server
	r := gin.Default()
	// Define routes and handlers
	r.POST("/api/v1/users", uHandler.RegisterUser)
	r.GET("/api/v1/users/:id", uHandler.GetUserProfile)
	r.PUT("/api/v1/users/:id", uHandler.UpdateUserProfile)

	// Start the server
	logger.Log.Info("Server running on port 8080")
	if err := r.Run(":8080"); err != nil {
		logger.Log.Error("Failed to start the server: %v", err)
	}

}
