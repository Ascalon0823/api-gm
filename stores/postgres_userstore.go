package stores

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	// Register the PostgreSQL dialector
	RegisterDialector("postgres", func(dsn string) gorm.Dialector {
		return postgres.Open(dsn)
	})
}
