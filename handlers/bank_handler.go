package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var BankCollection *mongo.Collection

func InitBankCollection(collection *mongo.Collection) {
	BankCollection = collection

	// Ensure unique compound index on bankName + code
	_, err := BankCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "bankName", Value: 1},
				{Key: "code", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Println("⚠️ Could not create index:", err)
	}
}

func SearchBanks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := r.URL.Query().Get("query")
	log.Println("🔍 Search query:", query)

	// 1. Check Mongo first
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
			log.Println("✅ Found in Mongo:", len(results))
			json.NewEncoder(w).Encode(results)
			return
		}
	}
	if cursor != nil {
		cursor.Close(ctx)
	}

	// 2. Not in Mongo → fetch from Arca
	log.Println("🌍 Fetching from Arca...")
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

	log.Println("📦 Arca response count:", len(banks))

	// 3. Normalize into your schema
	cleaned := make([]interface{}, 0, len(banks))
	for _, b := range banks {
		name, _ := b["bankName"].(string)
		code, _ := b["code"].(string) // ✅ use `code` directly
		if name == "" {
			continue
		}
		cleaned = append(cleaned, map[string]interface{}{
			"bankName": name,
			"code":     code,
			"id":       uuid.New().String(),
		})
	}

	log.Println("🧹 Normalized banks count:", len(cleaned))

	if len(cleaned) > 0 {
		_, err = BankCollection.InsertMany(ctx, cleaned)
		if err != nil {
			log.Println("⚠️ Insert error (duplicates likely skipped):", err)
		} else {
			log.Println("💾 Inserted into Mongo")
		}
	}

	// 4. Return filtered response to frontend
	if query != "" {
		filtered := []map[string]interface{}{}
		for _, c := range cleaned {
			row := c.(map[string]interface{})
			if rowName, ok := row["bankName"].(string); ok {
				if match, _ := regexp.MatchString("(?i)"+query, rowName); match {
					filtered = append(filtered, row)
				}
			}
		}
		log.Println("🔎 Filtered banks:", len(filtered))
		json.NewEncoder(w).Encode(filtered)
		return
	}

	json.NewEncoder(w).Encode(cleaned)
}
