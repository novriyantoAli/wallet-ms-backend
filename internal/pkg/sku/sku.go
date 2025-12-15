package sku

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Generator provides SKU generation and validation utilities
type Generator struct {
	prefix string
	maxLen int
}

// NewGenerator creates a new SKU generator with optional prefix
func NewGenerator(prefix string) *Generator {
	if prefix != "" && !isValidPrefix(prefix) {
		prefix = sanitizePrefix(prefix)
	}
	return &Generator{
		prefix: prefix,
		maxLen: 100, // Database constraint
	}
}

// Generate creates a unique SKU using timestamp and random suffix
// Format: [PREFIX-]TIMESTAMP-RANDOM
// Example: WIFI-1702556400000-AB3K9X or PULSA-1702556400000-XY7Z2M
func (g *Generator) Generate() (string, error) {
	timestamp := time.Now().UnixMilli()
	randomSuffix, err := generateRandomSuffix(6)
	if err != nil {
		return "", err
	}

	var sku string
	if g.prefix != "" {
		sku = fmt.Sprintf("%s-%d-%s", g.prefix, timestamp, randomSuffix)
	} else {
		sku = fmt.Sprintf("%d-%s", timestamp, randomSuffix)
	}

	// Validate length
	if len(sku) > g.maxLen {
		return "", fmt.Errorf("generated SKU exceeds maximum length of %d characters", g.maxLen)
	}

	return sku, nil
}

// GenerateWithPattern generates SKU with custom pattern
// Pattern examples:
// - WIFI-{timestamp}-{random}
// - PULSA-{sequence}
// - {prefix}-{date}-{random}
func (g *Generator) GenerateWithPattern(pattern string) (string, error) {
	sku := pattern
	sku = strings.ReplaceAll(sku, "{timestamp}", fmt.Sprintf("%d", time.Now().UnixMilli()))
	sku = strings.ReplaceAll(sku, "{date}", time.Now().Format("20060102"))
	sku = strings.ReplaceAll(sku, "{time}", time.Now().Format("150405"))

	if strings.Contains(sku, "{random}") {
		randomSuffix, err := generateRandomSuffix(6)
		if err != nil {
			return "", err
		}
		sku = strings.ReplaceAll(sku, "{random}", randomSuffix)
	}

	if strings.Contains(sku, "{prefix}") {
		sku = strings.ReplaceAll(sku, "{prefix}", g.prefix)
	}

	// Validate length and format
	if len(sku) > g.maxLen {
		return "", fmt.Errorf("generated SKU exceeds maximum length of %d characters", g.maxLen)
	}

	return sku, nil
}

// IsValid validates a SKU format
func (g *Generator) IsValid(skuValue string) bool {
	if skuValue == "" {
		return false
	}

	// Check length
	if len(skuValue) > g.maxLen {
		return false
	}

	// Check if it only contains alphanumeric and hyphens
	validPattern := regexp.MustCompile(`^[A-Za-z0-9\-]+$`)
	if !validPattern.MatchString(skuValue) {
		return false
	}

	// Check if prefix matches (if prefix is set)
	if g.prefix != "" && !strings.HasPrefix(skuValue, g.prefix+"-") {
		return false
	}

	return true
}

// ExtractPrefix extracts the prefix from a SKU
func ExtractPrefix(sku string) string {
	parts := strings.Split(sku, "-")
	if len(parts) > 1 {
		return parts[0]
	}
	return ""
}

// ExtractTimestamp extracts timestamp from generated SKU
// Returns 0 if SKU wasn't generated with timestamp pattern
func ExtractTimestamp(sku string) int64 {
	parts := strings.Split(sku, "-")
	if len(parts) < 2 {
		return 0
	}

	// Try to parse the second-to-last part as timestamp
	var timestampStr string
	if len(parts) == 3 {
		// Format: PREFIX-TIMESTAMP-RANDOM
		timestampStr = parts[1]
	} else if len(parts) == 2 {
		// Format: TIMESTAMP-RANDOM
		timestampStr = parts[0]
	}

	var timestamp int64
	_, err := fmt.Sscanf(timestampStr, "%d", &timestamp)
	if err != nil {
		return 0
	}

	return timestamp
}

// sanitizePrefix removes invalid characters from prefix
func sanitizePrefix(prefix string) string {
	// Only allow alphanumeric and underscores
	pattern := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	sanitized := pattern.ReplaceAllString(prefix, "")
	// Convert to uppercase
	return strings.ToUpper(sanitized)
}

// isValidPrefix checks if prefix format is valid
func isValidPrefix(prefix string) bool {
	validPattern := regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	return validPattern.MatchString(prefix)
}

// generateRandomSuffix generates a random alphanumeric string
func generateRandomSuffix(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random suffix: %w", err)
	}

	for i := range b {
		b[i] = charset[randomBytes[i]%byte(len(charset))]
	}

	return string(b), nil
}
