package env

import (
	"os"
	"sync"

	_ "github.com/lib/pq" // Importing pq for PostgreSQL driver
)

type Env struct {
	AddrClient string
	Addr       string
	DBHost     string
	DBPortOdoo string
	DBNameOdoo string
	DBUserOdoo string
	DBPassOdoo string
	SSLMode    string
	SecretKey  string
}

var (
	cfg  *Env
	once sync.Once
)

func Start() *Env {
	once.Do(func() {
		cfg = &Env{
			AddrClient: getEnv("ADDR_CLIENT", "http://localhost:5173"),
			Addr:       getEnv("ADDR", "localhost:8060"),
			DBHost:     getEnv("DB_HOST", "localhost"),
			DBPortOdoo: getEnv("DB_PORT", "5433"),
			DBNameOdoo: getEnv("DB_NAME", "odoo"),
			DBUserOdoo: getEnv("DB_USER", "odoo"),
			DBPassOdoo: getEnv("DB_PASS", "odoo"),
			SSLMode:    getEnv("SSL_MODE", "disable"),
			SecretKey:  getEnv("SECRET_KEY", "mysecretkey"),
		}
	})
	return cfg

}

func getEnv(name string, fallback string) string {
	if env, ok := os.LookupEnv(name); ok {
		return env
	}
	return fallback

}
