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

	// init Elasticsearch client and indices
	if err := db.InitES(); err != nil {
		log.Fatalf("Failed to initialize Elasticsearch: %v", err)
	}

	// init short id generator
	if err := handlers.InitShortID(); err != nil {
		log.Fatalf("Failed to initialize short id generator: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	auth.Initialize(jwtSecret)
	authService := auth.NewService(jwtSecret)

	// Create handlers with dependencies
	voteHandler := handlers.NewVoteHandler()
	authHandler := handlers.NewAuthHandler(authService)

	r := mux.NewRouter()

	// Auth routes
	r.HandleFunc("/auth/register/initiate", authHandler.InitiateRegistration).Methods("POST")
	r.HandleFunc("/auth/register/complete", authHandler.CompleteRegistration).Methods("POST")

	r.HandleFunc("/auth/password/reset", authHandler.InitiatePasswordReset).Methods("POST")
	r.HandleFunc("/auth/password/complete", authHandler.CompletePasswordReset).Methods("POST")

	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	// Vote routes - protected by auth middleware
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(auth.AuthMiddleware)
	apiRouter.HandleFunc("/vote", voteHandler.HandleVote).Methods("POST")
	apiRouter.HandleFunc("/votes/count", voteHandler.GetVoteCount).Methods("GET")

	rankingHandler := handlers.NewRankingHandler()
	apiRouter.HandleFunc("/rankings", rankingHandler.CreateRanking).Methods("POST")
	apiRouter.HandleFunc("/rankings", rankingHandler.UpdateRanking).Methods("PATCH")
	apiRouter.HandleFunc("/rankings", rankingHandler.GetUserRankings).Methods("GET")
	apiRouter.HandleFunc("/rankings", rankingHandler.GetUserRankings).Methods("GET")
	apiRouter.HandleFunc("/rankings/{rankingID}", rankingHandler.GetRanking).Methods("GET")
	apiRouter.HandleFunc("/rankings/{rankingID}", rankingHandler.DeleteRanking).Methods("DELETE")

	port := getPort()

	// Start cleanup goroutine for unconfirmed users
	go auth.StartCleanupJob()

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getPort() string {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
