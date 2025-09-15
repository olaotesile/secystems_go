// models/bank.go
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Bank struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	BankName string             `json:"bankName" bson:"name"`
	Code     string             `json:"code" bson:"code"`
	LogoURL  string             `json:"logoUrl" bson:"logo"`
}