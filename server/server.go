package server

import (
	"cmd/middleware"
	"cmd/stores"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func Start(userStore stores.UserStore) error {
	// Initialize the Gin router
	r := setupRouter(userStore)

	// Start the server on port 8080
	if err := r.Run(":8080"); err != nil {
		return err
	}
	return nil
}

func getJwtSecret() string {
	return os.Getenv("JWT_SECRET")
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func useSecureCookie() bool {
	return gin.Mode() == gin.ReleaseMode
}

func generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(getJwtSecret()))
}

func setupRouter(store stores.UserStore) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CorsMiddleware())
	r.POST("/register", handleRegister(store))
	r.POST("/login", handleLogin(store))
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware(getJwtSecret()))
	auth.GET("/me", handleMe(store))
	auth.POST("/logout", handleLogout())
	return r
}

func handleRegister(store stores.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		hashedPassword, err := hashPassword(req.Password)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to hash password"})
			return
		}
		req.Password = hashedPassword
		if user, _ := store.FindByEmail(c.Request.Context(), req.Email); user != nil {
			c.JSON(409, gin.H{"error": "User already exists"})
			return
		}
		user := stores.User{
			Email:     req.Email,
			Password:  hashedPassword,
			CreatedAt: time.Now(),
		}
		if err := store.Create(c.Request.Context(), &user); err != nil {
			c.JSON(500, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(201, gin.H{"message": "User created successfully"})
	}
}

func handleLogin(store stores.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		storedUser, err := store.FindByEmail(c.Request.Context(), req.Email)
		if err != nil {
			println(req.Email, "not found in database")
			c.JSON(401, gin.H{"error": "User not found or invalid credentials"})
			return
		}
		if err := checkPassword(storedUser.Password, req.Password); err != nil {
			println("Password mismatch for user:", req.Email)
			c.JSON(401, gin.H{"error": "User not found or invalid credentials"})
			return
		}
		tokenString, err := generateToken(storedUser.ID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create token"})
			return
		}
		c.SetCookie("token", tokenString, 3600, "/", "", useSecureCookie(), true)
		c.JSON(200, gin.H{"message": "Login successful"})
	}
}

func handleMe(store stores.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		storedUser, err := store.FindByID(c.Request.Context(), userID.(uint))
		if err != nil {
			c.JSON(404, gin.H{"error": "User not found"})
			return
		}
		storedUser.Password = "" // Don't return the password
		c.JSON(200, storedUser)
	}
}

func handleLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear the token cookie
		c.SetCookie("token", "", -1, "/", "", useSecureCookie(), true)
		c.JSON(200, gin.H{"message": "Logged out successfully"})
	}
}
