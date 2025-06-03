package user

import (
	"cmd/email"
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
			Email:    req.Email,
			Password: hashedPassword,
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

func handleForgotPassword(store UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		storedUser, err := store.FindByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(200, gin.H{"message": "Password reset link sent to your email if the account exists"})
			return
		}
		storedUser.ResetToken, err = generateResetToken()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate reset token"})
			return
		} // This should be a generated token
		storedUser.ResetExpiry = time.Now().Add(24 * time.Hour) // Set expiry for 24 hours
		if err := store.Update(c.Request.Context(), storedUser); err != nil {
			c.JSON(500, gin.H{"error": "Failed to update user for password reset"})
			return
		}
		println("Forgot password request for email:", storedUser.Email, "with reset token:", storedUser.ResetToken)
		if err := email.SendResetEmail(req.Email, "123"); err != nil {
			println("Failed to send reset email:", err)
			c.JSON(500, gin.H{"error": "Failed to send password reset email"})
			return
		}

		c.JSON(200, gin.H{"message": "Password reset link sent to your email if the account exists"})
	}
}
func handlePasswordReset(store UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email              string `json:"email"`
			PasswordResetToken string `json:"password_reset_token"`
			NewPassword        string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		if req.PasswordResetToken == "" || req.NewPassword == "" {
			c.JSON(400, gin.H{"error": "Invalid input - token and new password are required"})
			return
		}
		storedUser, err := store.FindByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid input - user not found"})
			return
		}
		if storedUser.ResetToken != req.PasswordResetToken {
			c.JSON(400, gin.H{"error": "Invalid input - token does not match"})
			return
		}
		if time.Now().After(storedUser.ResetExpiry) {
			c.JSON(400, gin.H{"error": "Invalid input - token has expired"})
			return
		}
		storedUser.ResetToken = ""           // Clear the reset token after use
		storedUser.ResetExpiry = time.Time{} // Clear the expiry time
		println("Resetting password for user:", req.Email, "with token:", req.PasswordResetToken)
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

		c.JSON(200, gin.H{"message": "Password reset successfully"})
	}
}
