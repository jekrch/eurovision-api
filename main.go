package main

import (
	"eurovision-api/auth"
	"eurovision-api/db"
	"eurovision-api/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {

	// Initialize Elasticsearch
	if err := db.InitES(); err != nil {
		log.Fatalf("Failed to initialize Elasticsearch: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-default-secret-key" // for development only
		log.Println("Warning: Using default JWT secret")
	}

	authService := auth.NewService(jwtSecret)

	// Create handlers with dependencies
	voteHandler := handlers.NewVoteHandler()
	authHandler := handlers.NewAuthHandler(authService)

	r := mux.NewRouter()

	// Auth routes
	r.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/auth/confirm", authHandler.Confirm).Methods("GET")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	// Vote routes - protected by auth middleware
	voteRouter := r.PathPrefix("/api").Subrouter()
	voteRouter.Use(auth.AuthMiddleware)
	voteRouter.HandleFunc("/vote", voteHandler.HandleVote).Methods("POST")
	voteRouter.HandleFunc("/votes/count", voteHandler.GetVoteCount).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start cleanup goroutine for unconfirmed users
	go auth.StartCleanupJob()

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
