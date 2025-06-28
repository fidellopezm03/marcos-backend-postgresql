package main

import (
	"log"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/api"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/handler"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/internal/db"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/internal/env"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/repository"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {

	env := env.Start()

	api := api.NewApi(env.Addr)
	connOdoo := db.GetConnectionOdoo(db.DBConfig{
		Host:     env.DBHost,
		Port:     env.DBPortOdoo,
		User:     env.DBUserOdoo,
		Password: env.DBPassOdoo,
		Name:     env.DBNameOdoo,
		SSLMode:  env.SSLMode,
	})
	connAdmin := db.GetConnectionAdmin(db.DBConfig{
		Host:     env.DBHost,
		Port:     env.DBPortAdmin,
		User:     env.DBUserAdmin,
		Password: env.DBPassAdmin,
		Name:     env.DBNameAdmin,
		SSLMode:  env.SSLMode,
	})

	defer func() {
		connOdoo.Close()
		connAdmin.Close()
	}()

	repositoryOdoo := repository.NewProductRepo(connOdoo)
	repositoryAdmin := repository.NewAdminRepo(connAdmin)

	productService := service.NewProductService(repositoryOdoo)

	productHandlerOdoo := handler.NewProductHandler(productService)
	adminHandler := handler.NewAdminHandler(repositoryAdmin)

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.AddrClient},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	productHandlerOdoo.RegisterRoutes(router)
	adminHandler.RegisterRoutes(router)
	// newHashedPassword, err := bcrypt.GenerateFromPassword([]byte("adminmaitepassword"), bcrypt.DefaultCost)
	// if err == nil {
	// 	log.Println(string(newHashedPassword))
	// }

	log.Fatal("Error in server ", api.Run(router))

}
