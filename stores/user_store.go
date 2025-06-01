package stores

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type User struct {
	ID          uint   `json:"id" bson:"_id,omitempty" gorm:"primaryKey"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email" gorm:"uniqueIndex"`
	Password    string `json:"password"`
	CreatedAt   time.Time
}

type UserStore interface {
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
}

func NewUserStore() (UserStore, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to environment variables")
		return nil, err
	}
	dbDialect := os.Getenv("DB_DIALECT")
	dsn := os.Getenv("DB_DSN")
	gormDB, err := NewGormDB(dbDialect, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}
	return NewGormUserStore(gormDB), nil
}
