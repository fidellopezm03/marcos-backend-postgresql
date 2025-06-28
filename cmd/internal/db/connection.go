package db

import (
	"database/sql"
	"fmt"
	"log"
)

var (
	dbOdoo  *sql.DB
	dbAdmin *sql.DB
)

type DBConfig struct {
	Host string
	Port string

	User     string
	Password string
	Name     string
	SSLMode  string
}

func GetConnectionOdoo(config DBConfig) *sql.DB {
	return getConnection(config, dbOdoo)
}
func GetConnectionAdmin(config DBConfig) *sql.DB {
	return getConnection(config, dbAdmin)
}

func getConnection(config DBConfig, sqlv *sql.DB) *sql.DB {
	if sqlv == nil {
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.Name, config.SSLMode)

		var err error
		sqlv, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Error connecting to the database: %v", err)
		}

		if err = sqlv.Ping(); err != nil {
			log.Fatalf("Error pinging the database: %v", err)
		}
		log.Printf("Database connection established successfully: %s", config.Name)
	}
	return sqlv
}
