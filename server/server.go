package server

import (
	"cmd/internal/user"
	"cmd/middleware"

	"github.com/gin-gonic/gin"
)

func Start() error {
	// Initialize the Gin router
	r := gin.Default()
	r.Use(middleware.CorsMiddleware())
	if err := user.Start(r); err != nil {
		println("Error starting user service:", err)
		return err
	}
	// Start the server on port 8080
	if err := r.Run(":8080"); err != nil {
		return err
	}
	return nil
}
