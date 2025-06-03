package stores

import (
	"gorm.io/gorm"
)

type DialectorFunc func(dsn string) gorm.Dialector

var dialectorRegistry = make(map[string]DialectorFunc)
var connectionRegistry = make(map[string]*gorm.DB)

func RegisterDialector(name string, dialectorFunc DialectorFunc) {
	dialectorRegistry[name] = dialectorFunc
}

func NewGormDB(dialect, dsn string) (*gorm.DB, error) {
	connection, exists := connectionRegistry[dialect]
	if exists {
		return connection, nil
	}
	dialectorFunc, exists := dialectorRegistry[dialect]
	if !exists {
		return nil, gorm.ErrUnsupportedDriver
	}
	dialector := dialectorFunc(dsn)
	connection, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	connectionRegistry[dialect] = connection
	return connection, nil
}
