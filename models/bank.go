package models

type Bank struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	BankName    string `json:"bankName" bson:"bankName"`
	Shortcode   string `json:"shortcode" bson:"shortcode"`
	LogoURL     string `json:"logoUrl" bson:"logoUrl"`
	LastUpdated string `json:"lastUpdated" bson:"lastUpdated"`
}