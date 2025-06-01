package stores

import (
	"context"

	"gorm.io/gorm"
)

type GormUserStore struct {
	db *gorm.DB
}
type DialectorFunc func(dsn string) gorm.Dialector

var dialectorRegistry = make(map[string]DialectorFunc)

func RegisterDialector(name string, dialectorFunc DialectorFunc) {
	dialectorRegistry[name] = dialectorFunc
}

func NewGormDB(dialect, dsn string) (*gorm.DB, error) {
	dialectorFunc, exists := dialectorRegistry[dialect]
	if !exists {
		return nil, gorm.ErrUnsupportedDriver
	}
	dialector := dialectorFunc(dsn)
	return gorm.Open(dialector, &gorm.Config{})
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
