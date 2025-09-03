package middleware

import (
	"log"
	"net/http"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/internal/env"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

func RecoverPanic() func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("panic recovered: %v", rec)
					render.Status(r, http.StatusInternalServerError)
					render.JSON(w, r, map[string]string{"error": "internal server error"})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
func CORSmiddleware(env *env.Env) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.AddrClient},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
