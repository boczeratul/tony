package db

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB instantiates sql client
func InitDB(dbHost, dbName, dbPort, dbUser, dbPassword string) (*gorm.DB, error) {
	// connect to db
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open DB with DSN %s: %v", dsn, err)
	}

	pool, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %v", err)
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	pool.SetMaxIdleConns(25)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	pool.SetMaxOpenConns(25)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	pool.SetConnMaxLifetime(10 * time.Minute)

	return db, nil
}
