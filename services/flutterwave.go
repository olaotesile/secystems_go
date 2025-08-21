package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"secsystems-go/models"
)

type FlutterwaveBank struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type FlutterwaveResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Data    []FlutterwaveBank `json:"data"`
}

func GetBanksFromFlutterwave() ([]models.Bank, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	clientID := os.Getenv("FLUTTERWAVE_CLIENT_ID")
	clientSecret := os.Getenv("FLUTTERWAVE_CLIENT_SECRET")

	
	tokenURL := "https://idp.flutterwave.com/realms/flutterwave/protocol/openid-connect/token"
	data := fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=client_credentials", clientID, clientSecret)

	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(body, &tokenResp)


	banksURL := "https://api.flutterwave.com/v3/banks/NG"
	req, _ = http.NewRequest("GET", banksURL, nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	var fwResp FlutterwaveResponse
	json.Unmarshal(body, &fwResp)

	
	var banks []models.Bank
now := time.Now().UTC().Format(time.RFC3339) 

for _, b := range fwResp.Data {
	banks = append(banks, models.Bank{
		BankName:    b.Name,
		Shortcode:   b.Code,
		LogoURL:     b.Logo,
		LastUpdated: now, 
	})
}

	return banks, nil
}