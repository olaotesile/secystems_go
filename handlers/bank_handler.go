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
	"go.mongodb.org/mongo-driver/mongo/options"

	"secsystems-go/models"   
	"secsystems-go/services" 
)


var BankCollection *mongo.Collection


func SearchBanks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}


	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cachedBanks []models.Bank
	filter := bson.M{
		"$or": []bson.M{
			{"bankName": bson.M{"$regex": query, "$options": "i"}},
			{"shortcode": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	cursor, err := BankCollection.Find(ctx, filter)
	if err != nil {
		log.Println("❌ Error querying MongoDB:", err)
	} else {
		if err = cursor.All(ctx, &cachedBanks); err == nil && len(cachedBanks) > 0 {
			response := map[string]interface{}{
				"success": true,
				"data":    cachedBanks,
				"message": "Banks fetched from cache",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	
	banks, err := services.GetBanksFromFlutterwave()
	if err != nil {
		http.Error(w, "Failed to fetch banks from Flutterwave", http.StatusInternalServerError)
		return
	}

	
	var filtered []models.Bank
	for _, b := range banks {
		if strings.Contains(strings.ToLower(b.BankName), strings.ToLower(query)) ||
			strings.Contains(b.Shortcode, query) {
			filtered = append(filtered, b)
		}
	}

	
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var docs []interface{}
		for _, b := range banks {
			docs = append(docs, b)
		}

	
		for _, doc := range docs {
			_, err := BankCollection.ReplaceOne(
				ctx,
				bson.M{"shortcode": doc.(models.Bank).Shortcode},
				doc,
				options.Replace().SetUpsert(true),
			)
			if err != nil {
				log.Println("❌ Failed to save bank to MongoDB:", err)
			}
		}
	}()

	
	response := map[string]interface{}{
		"success": true,
		"data":    filtered,
		"message": "Banks fetched from Flutterwave and cached",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}