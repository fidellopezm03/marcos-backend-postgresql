package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)
var (
	db *sql.DB
	once sync.Once
)
type DBConfig struct {
	Host     string
	Port     string
	
	User     string
	Password string
	Name     string
	SSLMode  string
}
func GetConnection(config DBConfig) *sql.DB {
	once.Do(func() {
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.Name, config.SSLMode)
		
		var err error
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Error connecting to the database: %v", err)
		}
		
		if err = db.Ping(); err != nil {
			log.Fatalf("Error pinging the database: %v", err)
		}
		log.Println("Database connection established successfully")
	})
	return db
}