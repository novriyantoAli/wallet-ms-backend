package util

import (
	"fmt"
	"strings"
	"unicode"
)

// GenerateSKU generates a unique SKU from category, name, and price
// Format: {CATEGORY_PREFIX}-{NAME_HASH}-{PRICE_HASH}
// Example: WIFI-A2B5-25K
func GenerateSKU(category, name string, price float64) string {
	// Create category prefix (first 4 chars uppercase)
	categoryPrefix := strings.ToUpper(category)
	if len(categoryPrefix) > 4 {
		categoryPrefix = categoryPrefix[:4]
	}

	// Create name hash (first 4 chars of MD5, uppercase)
	nameHash := generateHash(strings.ToUpper(name))

	// Create price hash based on price value
	priceHash := generatePriceHash(price)

	return fmt.Sprintf("%s-%s-%s", categoryPrefix, nameHash, priceHash)
}

// generateHash creates a 4-character hash from a string
func generateHash(input string) string {
	// Remove spaces and special chars, keep only alphanumeric
	cleaned := ""
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cleaned += string(r)
		}
	}

	if len(cleaned) == 0 {
		cleaned = "PROD"
	}

	// Use first 4 characters if available, otherwise pad
	if len(cleaned) >= 4 {
		return cleaned[:4]
	}
	return fmt.Sprintf("%-4s", cleaned)
}

// generatePriceHash creates a hash from the price
// Uses price brackets for simplicity: 1K, 5K, 10K, 50K, 100K, etc.
func generatePriceHash(price float64) string {
	if price < 0 {
		return "FREE"
	}

	switch {
	case price <= 1000:
		return "1K"
	case price <= 5000:
		return "5K"
	case price <= 10000:
		return "10K"
	case price <= 50000:
		return "50K"
	case price <= 100000:
		return "100K"
	case price <= 500000:
		return "500K"
	default:
		return "1M+"
	}
}
