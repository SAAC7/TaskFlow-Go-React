package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"taskflow-backend/internal/config"
	"taskflow-backend/internal/handlers"
	"taskflow-backend/internal/repositories"
	"taskflow-backend/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Connect to MongoDB
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Initialize layers
	taskRepository := repositories.NewTaskRepository(db)
	taskService := services.NewTaskService(taskRepository)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Configure routes
	mux := http.NewServeMux()

	mux.HandleFunc("POST /tasks", taskHandler.Create)
	mux.HandleFunc("GET /tasks", taskHandler.FindAll)
	mux.HandleFunc("GET /tasks/{id}", taskHandler.FindByID)
	mux.HandleFunc("PUT /tasks/{id}", taskHandler.Update)
	mux.HandleFunc("DELETE /tasks/{id}", taskHandler.Delete)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Running Go Server on http://localhost:" + port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println(err)
	}
}
