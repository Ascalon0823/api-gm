package user

import (
	"cmd/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, store UserStore) *gin.Engine {
	r.Use(middleware.CorsMiddleware())
	r.POST("/register", handleRegister(store))
	r.POST("/login", handleLogin(store))
	r.POST("/password-forgot", handleForgotPassword(store))
	r.POST("/password-reset", handlePasswordReset(store))
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware(getJwtSecret()))
	auth.GET("/me", handleMe(store))
	auth.POST("/logout", handleLogout())
	auth.POST("/change-password", handleChangePassword(store))
	return r
}
