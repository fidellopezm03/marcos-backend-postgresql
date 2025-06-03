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
)


func main(){
	
	env := env.Start()
	
	api := api.NewApi(env.Addr)
	conn := db.GetConnection(db.DBConfig{
		Host:     env.DBHost,
		Port:     env.DBPort,
		User:     env.DBUser,
		Password: env.DBPass,
		Name:     env.DBName,
		SSLMode:  env.SSLMode,
	})
	
	defer conn.Close()
	

	repository:=repository.NewProductRepo(conn)
	productService:=service.NewProductService(repository)
	productHandler:=handler.NewProductHandler(productService)
	router := chi.NewRouter()
	productHandler.RegisterRoutes(router)
	
	
	
	log.Fatal("Error in server ",api.Run(router))
	
}
