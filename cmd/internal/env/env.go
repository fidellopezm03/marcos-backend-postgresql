package env

import (
	"os"
	"sync"

	_ "github.com/lib/pq" // Importing pq for PostgreSQL driver
)

type Env struct {
	Addr string
	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string
	SSLMode string
	SecretKey string
	
}

var (cfg *Env 
	once sync.Once)
	

func Start() *Env{
	once.Do(func() {
	cfg = &Env{
		Addr: getEnv("ADDR", "localhost:8060"),
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5433"),
		DBName: getEnv("DB_NAME", "odoo"),
		DBUser: getEnv("DB_USER", "odoo"),
		DBPass: getEnv("DB_PASS", "odoo"),
		SSLMode: getEnv("SSL_MODE", "disable"),
		SecretKey: getEnv("SECRET_KEY","mysecretkey"),
		

	}})
	return cfg

}

func getEnv(name string, fallback string) string {
	if env,ok := os.LookupEnv(name); ok{
		return env
	}
	return fallback

}