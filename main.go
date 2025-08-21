package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"secsystems-go/handlers"
)

var Client *mongo.Client
var BankCollection *mongo.Collection

func main() {
	fmt.Println("🚀 Secsystems Go Backend Starting...")
	clientOptions := options.Client().ApplyURI("mongodb+srv://bootesile:9Pcl8yhdJquOK8Ec@cluster0.y33atcb.mongodb.net/secsystems?retryWrites=true&w=majority&appName=Cluster0")
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

	
	http.HandleFunc("/banks", handlers.SearchBanks)
	fmt.Println("✅ Server is running on http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}