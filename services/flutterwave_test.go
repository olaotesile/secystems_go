// services/flutterwave_test.go
package services

import (
	"testing"
	"secsystems-go/models"
)

// TestGetBanksFromFlutterwave tests the core function
func TestGetBanksFromFlutterwave(t *testing.T) {
	// Call the function
	banks, err := GetBanksFromFlutterwave()

	// Check for error
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if we got banks
	if len(banks) == 0 {
		t.Fatal("Expected at least one bank, got none")
	}

	// Check each bank has required fields
	for _, bank := range banks {
		if bank.BankName == "" {
			t.Error("Bank name is empty")
		}
		if bank.Shortcode == "" {
			t.Error("Shortcode is empty")
		}
		// LogoURL can be empty — optional
	}
}