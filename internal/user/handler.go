package user

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func handleRegister(store UserStore) gin.HandlerFunc {
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
		user := User{
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

func handleLogin(store UserStore) gin.HandlerFunc {
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

func handleMe(store UserStore) gin.HandlerFunc {
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

func handleChangePassword(store UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		println("Changing password for user ID:", userID.(uint))
		storedUser, err := store.FindByID(c.Request.Context(), userID.(uint))
		if err != nil {
			c.JSON(404, gin.H{"error": "User not found"})
			return
		}

		if err := checkPassword(storedUser.Password, req.OldPassword); err != nil {
			c.JSON(401, gin.H{"error": "Old password is incorrect"})
			return
		}

		hashedNewPassword, err := hashPassword(req.NewPassword)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to hash new password"})
			return
		}

		storedUser.Password = hashedNewPassword
		if err := store.Update(c.Request.Context(), storedUser); err != nil {
			c.JSON(500, gin.H{"error": "Failed to update password"})
			return
		}

		c.JSON(200, gin.H{"message": "Password changed successfully"})
	}
}
