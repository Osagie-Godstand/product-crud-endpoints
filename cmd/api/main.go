package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Osagie-Godstand/product-crud-endpoints/db"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	config := &db.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	dbConn, err := db.NewConnection(config)
	if err != nil {
		log.Fatal("could not connect to the database:", err)
	}

	migrationsErr := db.RunMigrations(dbConn)
	if migrationsErr != nil {
		log.Fatal("could not migrate the database:", migrationsErr)
	}

	createNewProducts(dbConn)

	r := &ProductStore{
		DB: dbConn,
	}

	app := chi.NewRouter()
	r.setupRoutes(app)
	listenAddr := os.Getenv("HTTP_LISTEN_ADDRESS")
	fmt.Printf("Server listening on %s\n", listenAddr)
	if err := http.ListenAndServe(listenAddr, app); err != nil {
		log.Fatalf("Server failed to start: %s", err)
	}
}
