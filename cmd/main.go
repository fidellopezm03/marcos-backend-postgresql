package main

import (
	"log"
	"os"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/api"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/handler"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/internal/db"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/internal/env"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/repository"
	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	log.Println("Starting server...")
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
	log.Println("Database connection successful")

	if err := db.ApplyMigrations(connOdoo, "cmd/internal/db/migration.sql"); err != nil {
		log.Fatalf("error applying migrations: %v", err)
	}
	log.Println("Migrations applied successfully")
	defer connOdoo.Close()

	if err := os.MkdirAll(repository.UploadDir, 0755); err != nil {
		log.Fatalf("error detecting file/img: %v", err)
	}

	repositoryOdoo := repository.NewProductRepo(connOdoo)
	repositoryAdmin := repository.NewAdminRepo(connOdoo)

	productService := service.NewProductService(repositoryOdoo)

	productHandlerOdoo := handler.NewProductHandler(productService)
	adminHandler := handler.NewAdminHandler(repositoryAdmin)

	router := chi.NewRouter()

	productHandlerOdoo.RegisterRoutes(router, env)
	adminHandler.RegisterRoutes(router)

	log.Fatal("Error in server ", api.Run(router))

}
