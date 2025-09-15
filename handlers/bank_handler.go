// handlers/bank_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"secsystems-go/models"
)

var BankCollection *mongo.Collection

// InitBankCollection initializes the MongoDB collection.
// Note: Removed index on "bankName" → should be "name" (actual DB field)
func InitBankCollection(collection *mongo.Collection) {
	BankCollection = collection

	// Create unique index on ACTUAL DB fields: "name" + "code"
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},   // ← Use "name", not "bankName"
			{Key: "code", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	_, err := BankCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Println("⚠️ Could not create unique index:", err)
	} else {
		log.Println("✅ Unique index created on name + code")
	}
}

// SearchBanks searches for banks in MongoDB only (no external APIs)
func SearchBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := r.URL.Query().Get("query")
	log.Println("🔍 Search query received:", query)

	filter := bson.M{}
	if query != "" {
		// Case-insensitive regex search on actual DB field: "name"
		filter["name"] = bson.M{"$regex": regexp.QuoteMeta(query), "$options": "i"}
	}

	// Always initialize as empty slice to return [] not null
	banks := []models.Bank{}

	cursor, err := BankCollection.Find(ctx, filter)
	if err != nil {
		log.Printf("❌ Failed to query database: %v", err)
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// This fails if models.Bank doesn't match DB structure exactly
	if err = cursor.All(ctx, &banks); err != nil {
		log.Printf("❌ Failed to decode results: %v", err)
		http.Error(w, "Failed to decode database results", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Found %d banks matching query '%s'", len(banks), query)
	json.NewEncoder(w).Encode(banks)
}