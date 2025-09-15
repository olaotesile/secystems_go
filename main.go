// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	localHandlers "secsystems-go/handlers"
)

func main() {
	fmt.Println("🚀 Secsystems Go Backend Starting...")

	// Load .env if present (for local dev), but don't fail if missing
	if err := godotenv.Load(); err != nil {
		log.Println("🟡 No .env file found, using environment variables")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("❌ MONGODB_URI is not set in environment")
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("❌ Failed to connect to MongoDB:", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ Failed to ping MongoDB:", err)
	}

	fmt.Println("✅ Connected to MongoDB Atlas!")

	// Initialize the bank collection
	bankCollection := client.Database("secsystems").Collection("bankmappings")
	localHandlers.InitBankCollection(bankCollection) // This also sets up indexes

	// Set port
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Setup CORS (fixed: removed extra spaces in origins)
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{
			"http://localhost:5173",
			"https://secsystems-frontend.vercel.app", // Removed trailing space
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)(http.HandlerFunc(localHandlers.SearchBanks))

	// Register routes
	http.Handle("/banks", corsHandler)

	// Optional health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Printf("✅ Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}