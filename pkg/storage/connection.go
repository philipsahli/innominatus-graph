package storage

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeSQLite   DatabaseType = "sqlite"
)

type Config struct {
	Type     DatabaseType // "postgres" or "sqlite"
	Host     string       // PostgreSQL only
	Port     int          // PostgreSQL only
	User     string       // PostgreSQL only
	Password string       // PostgreSQL only
	DBName   string       // Database name or SQLite file path
	SSLMode  string       // PostgreSQL only
}

// NewConnection creates a database connection based on the configuration type
func NewConnection(config Config) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	var db *gorm.DB
	var err error

	switch config.Type {
	case DatabaseTypeSQLite:
		db, err = gorm.Open(sqlite.Open(config.DBName), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
		}
	case DatabaseTypePostgres:
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	return db, nil
}

// NewPostgresConnection creates a PostgreSQL connection (convenience function)
func NewPostgresConnection(host, user, password, dbname, sslmode string, port int) (*gorm.DB, error) {
	return NewConnection(Config{
		Type:     DatabaseTypePostgres,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  sslmode,
	})
}

// NewSQLiteConnection creates a SQLite connection (convenience function)
func NewSQLiteConnection(filepath string) (*gorm.DB, error) {
	return NewConnection(Config{
		Type:   DatabaseTypeSQLite,
		DBName: filepath,
	})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&App{}, &NodeModel{}, &EdgeModel{}, &GraphRunModel{})
}