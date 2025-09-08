package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var BankCollection *mongo.Collection // ensure this is set in main.go

func SearchBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Fetch bank codes from Arca API
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://arcapos-pay-middleware.qa.arca-payments.network/v9/pwts/cgate/ussd/bankcodes")
	if err != nil {
		http.Error(w, "Failed to fetch banks", http.StatusInternalServerError)
		log.Println("❌ Error fetching from Arca:", err)
		return
	}
	defer resp.Body.Close()

	var banks []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&banks); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusInternalServerError)
		log.Println("❌ Error decoding JSON:", err)
		return
	}

	// Filter out entries with nil internalCode and remove duplicates
	cleaned := make([]interface{}, 0, len(banks))
	seen := make(map[string]bool)
	for _, b := range banks {
		code, ok := b["internalCode"].(string)
		if !ok || code == "" {
			continue
		}
		if seen[code] {
			continue
		}
		seen[code] = true
		cleaned = append(cleaned, b)
	}

	// Store in MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Clear previous entries
	_, err = BankCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Println("❌ Error clearing previous banks:", err)
	}

	if len(cleaned) > 0 {
		_, err = BankCollection.InsertMany(ctx, cleaned)
		if err != nil {
			log.Println("❌ Error inserting banks into MongoDB:", err)
		} else {
			fmt.Println("✅ Bank codes stored in MongoDB")
		}
	}

	// Return the response to frontend
	json.NewEncoder(w).Encode(banks)
}
