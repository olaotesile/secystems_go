package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	localHandlers "secsystems-go/handlers"
	"github.com/gorilla/handlers"
)

func main() {
	fmt.Println("🚀 Secsystems Go Backend Starting...")

	
	/*
	clientOptions := options.Client().ApplyURI("your_mongo_uri")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ Failed to connect to MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("❌ Failed to ping MongoDB:", err)
	}

	fmt.Println("✅ Connected to MongoDB Atlas!")

	Client = client
	BankCollection = client.Database("secsystems").Collection("bankmappings")
	*/

	
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"https://secsystems-frontend.vercel.app"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
	)(http.HandlerFunc(localHandlers.SearchBanks))

	http.Handle("/banks", corsHandler)

	fmt.Printf("✅ Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
