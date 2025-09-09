package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var BankCollection *mongo.Collection

func SearchBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get query param
	query := r.URL.Query().Get("query")

	var filter bson.M
	if query != "" {
		filter = bson.M{
			"bankName": bson.M{
				"$regex":   query,
				"$options": "i", // case-insensitive
			},
		}
	} else {
		filter = bson.M{} // return all
	}

	cursor, err := BankCollection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to fetch banks", http.StatusInternalServerError)
		log.Println("❌ Error fetching banks:", err)
		return
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err = cursor.All(ctx, &results); err != nil {
		http.Error(w, "Failed to decode banks", http.StatusInternalServerError)
		log.Println("❌ Error decoding banks:", err)
		return
	}

	json.NewEncoder(w).Encode(results)
}
