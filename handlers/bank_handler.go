package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var BankCollection *mongo.Collection

func InitBankCollection(collection *mongo.Collection) {
	BankCollection = collection

	// Ensure unique index on bankName
	_, err := BankCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "bankName", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Println("⚠️ Could not create index on bankName:", err)
	}
}

func SearchBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := r.URL.Query().Get("query")

	// 1. Try Mongo first
	filter := bson.M{}
	if query != "" {
		filter["bankName"] = bson.M{
			"$regex":   query,
			"$options": "i",
		}
	}

	var results []map[string]interface{}
	cursor, err := BankCollection.Find(ctx, filter)
	if err == nil {
		if err = cursor.All(ctx, &results); err == nil && len(results) > 0 {
			json.NewEncoder(w).Encode(results)
			return
		}
	}
	if cursor != nil {
		cursor.Close(ctx)
	}

	// 2. Not in Mongo → fetch from Arca
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://arcapos-pay-middleware.qa.arca-payments.network/v9/pwts/cgate/ussd/bankcodes")
	if err != nil {
		http.Error(w, "Failed to fetch from Arca", http.StatusInternalServerError)
		log.Println("❌ Arca error:", err)
		return
	}
	defer resp.Body.Close()

	var banks []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&banks); err != nil {
		http.Error(w, "Failed to parse Arca response", http.StatusInternalServerError)
		log.Println("❌ JSON decode error:", err)
		return
	}

	// 3. Deduplicate before inserting
	seen := make(map[string]bool)
	cleaned := make([]interface{}, 0, len(banks))
	for _, b := range banks {
		if name, ok := b["bankName"].(string); ok && name != "" {
			if !seen[name] {
				seen[name] = true
				cleaned = append(cleaned, b)
			}
		}
	}

	if len(cleaned) > 0 {
		_, err = BankCollection.InsertMany(ctx, cleaned)
		if err != nil {
			log.Println("⚠️ Insert error (likely duplicates skipped):", err)
		}
	}

	// 4. After saving → re-run the same filter on Mongo
	cursor, err = BankCollection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to fetch banks after refresh", http.StatusInternalServerError)
		log.Println("❌ Error fetching after refresh:", err)
		return
	}
	defer cursor.Close(ctx)

	results = []map[string]interface{}{}
	if err = cursor.All(ctx, &results); err != nil {
		http.Error(w, "Failed to decode banks after refresh", http.StatusInternalServerError)
		log.Println("❌ Decode error after refresh:", err)
		return
	}

	json.NewEncoder(w).Encode(results)
}
