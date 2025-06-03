package user

import (
	"cmd/stores"
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

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
	gormDB, err := stores.NewGormDB(dbDialect, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}
	return NewGormUserStore(gormDB), nil
}

type GormUserStore struct {
	db *gorm.DB
}

func NewGormUserStore(db *gorm.DB) *GormUserStore {
	if err := db.AutoMigrate(&User{}); err != nil {
		panic("Failed to auto-migrate User model: " + err.Error())
	}
	return &GormUserStore{db: db}
}

func (s *GormUserStore) FindByID(ctx context.Context, id uint) (*User, error) {
	var user User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *GormUserStore) Create(ctx context.Context, user *User) error {
	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}
	return nil
}
func (s *GormUserStore) Update(ctx context.Context, user *User) error {
	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		return err
	}
	return nil
}
func (s *GormUserStore) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := s.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
