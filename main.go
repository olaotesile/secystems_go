package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	localHandlers "secsystems-go/handlers"
	"time"

	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("🚀 Secsystems Go Backend Starting...")

	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, relying on environment variables")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("❌ MONGODB_URI not set in environment")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("❌ Failed to connect to MongoDB:", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ Failed to ping MongoDB:", err)
	}

	fmt.Println("✅ Connected to MongoDB Atlas!")

	// Set the collection in the handler
	localHandlers.BankCollection = client.Database("secsystems").Collection("bankmappings")

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{
			"http://localhost:5173",
			"https://secsystems-frontend.vercel.app",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
	)(http.HandlerFunc(localHandlers.SearchBanks))

	http.Handle("/banks", corsHandler)

	fmt.Printf("✅ Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
