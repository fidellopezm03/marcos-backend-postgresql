package env

import (
	"os"
	"sync"

	_ "github.com/lib/pq" // Importing pq for PostgreSQL driver
)

type Env struct {
	AddrClient  string
	Addr        string
	DBHost      string
	DBPortOdoo  string
	DBPortAdmin string
	DBNameOdoo  string
	DBNameAdmin string
	DBUserOdoo  string
	DBUserAdmin string
	DBPassOdoo  string
	DBPassAdmin string
	SSLMode     string
	SecretKey   string
}

var (
	cfg  *Env
	once sync.Once
)

func Start() *Env {
	once.Do(func() {
		cfg = &Env{
			AddrClient:  getEnv("ADDR_CLIENT", "http://localhost:5173"),
			Addr:        getEnv("ADDR", "localhost:8060"),
			DBHost:      getEnv("DB_HOST", "localhost"),
			DBPortOdoo:  getEnv("DB_PORT", "5433"),
			DBPortAdmin: getEnv("DB_PORT_ADMIN", "5432"),
			DBNameOdoo:  getEnv("DB_NAME", "odoo"),
			DBNameAdmin: getEnv("DB_NAME_ADMIN", "tienda"),
			DBUserOdoo:  getEnv("DB_USER", "odoo"),
			DBUserAdmin: getEnv("DB_USER_ADMIN", "postgres"),
			DBPassOdoo:  getEnv("DB_PASS", "odoo"),
			DBPassAdmin: getEnv("DB_PASS_ADMIN", "postgres"),
			SSLMode:     getEnv("SSL_MODE", "disable"),
			SecretKey:   getEnv("SECRET_KEY", "mysecretkey"),
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
