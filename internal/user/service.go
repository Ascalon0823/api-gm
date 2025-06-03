package user

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func Start(r *gin.Engine) error {
	store, err := NewUserStore()
	if err != nil {
		println("Error initializing user store:", err)
		return err
	}
	SetupRouter(r, store)
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
