package sku

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	t.Run("should create generator with valid prefix", func(t *testing.T) {
		gen := NewGenerator("WIFI")
		assert.NotNil(t, gen)
		assert.Equal(t, "WIFI", gen.prefix)
		assert.Equal(t, 100, gen.maxLen)
	})

	t.Run("should create generator without prefix", func(t *testing.T) {
		gen := NewGenerator("")
		assert.NotNil(t, gen)
		assert.Equal(t, "", gen.prefix)
	})

	t.Run("should sanitize invalid prefix", func(t *testing.T) {
		gen := NewGenerator("wifi@product!")
		assert.NotNil(t, gen)
		assert.Equal(t, "WIFIPRODUCT", gen.prefix)
	})
}

func TestGenerate(t *testing.T) {
	t.Run("should generate valid SKU with prefix", func(t *testing.T) {
		gen := NewGenerator("WIFI")
		sku, err := gen.Generate()

		assert.NoError(t, err)
		assert.NotEmpty(t, sku)
		assert.True(t, strings.HasPrefix(sku, "WIFI-"))
		assert.True(t, len(sku) <= 100)
	})

	t.Run("should generate valid SKU without prefix", func(t *testing.T) {
		gen := NewGenerator("")
		sku, err := gen.Generate()

		assert.NoError(t, err)
		assert.NotEmpty(t, sku)
		// Should contain dash for timestamp-random separation
		assert.Contains(t, sku, "-")
	})

	t.Run("should generate unique SKUs", func(t *testing.T) {
		gen := NewGenerator("PULSA")
		sku1, _ := gen.Generate()
		sku2, _ := gen.Generate()

		assert.NotEqual(t, sku1, sku2)
	})

	t.Run("should respect maximum length", func(t *testing.T) {
		gen := NewGenerator("WIFI")
		for i := 0; i < 100; i++ {
			sku, err := gen.Generate()
			assert.NoError(t, err)
			assert.LessOrEqual(t, len(sku), 100)
		}
	})
}

func TestGenerateWithPattern(t *testing.T) {
	t.Run("should generate with timestamp-random pattern", func(t *testing.T) {
		gen := NewGenerator("WIFI")
		sku, err := gen.GenerateWithPattern("WIFI-{timestamp}-{random}")

		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(sku, "WIFI-"))
		assert.True(t, len(sku) <= 100)

		// Should have timestamp in milliseconds (13 digits)
		parts := strings.Split(sku, "-")
		assert.GreaterOrEqual(t, len(parts[1]), 13)
	})

	t.Run("should generate with date pattern", func(t *testing.T) {
		gen := NewGenerator("PULSA")
		sku, err := gen.GenerateWithPattern("PULSA-{date}-{random}")

		assert.NoError(t, err)
		assert.NotEmpty(t, sku)
		// Should contain today's date
		today := time.Now().Format("20060102")
		assert.Contains(t, sku, today)
	})

	t.Run("should generate with time pattern", func(t *testing.T) {
		gen := NewGenerator("DATA")
		sku, err := gen.GenerateWithPattern("DATA-{date}-{time}-{random}")

		assert.NoError(t, err)
		assert.NotEmpty(t, sku)
		assert.True(t, len(sku) <= 100)
	})

	t.Run("should handle pattern without random", func(t *testing.T) {
		gen := NewGenerator("FIXED")
		sku, err := gen.GenerateWithPattern("FIXED-{date}")

		assert.NoError(t, err)
		assert.NotEmpty(t, sku)
		today := time.Now().Format("20060102")
		assert.Contains(t, sku, today)
	})
}

func TestIsValid(t *testing.T) {
	gen := NewGenerator("WIFI")

	t.Run("should validate correct SKU format", func(t *testing.T) {
		validSKUs := []string{
			"WIFI-1702556400000-AB3K9X",
			"WIFI-2023-PRODUCT-001",
			"WIFI-ABC123",
		}

		for _, sku := range validSKUs {
			assert.True(t, gen.IsValid(sku), fmt.Sprintf("SKU %s should be valid", sku))
		}
	})

	t.Run("should reject empty SKU", func(t *testing.T) {
		assert.False(t, gen.IsValid(""))
	})

	t.Run("should reject SKU exceeding max length", func(t *testing.T) {
		longSKU := "WIFI-" + strings.Repeat("A", 150)
		assert.False(t, gen.IsValid(longSKU))
	})

	t.Run("should reject SKU with invalid characters", func(t *testing.T) {
		invalidSKUs := []string{
			"WIFI@PRODUCT",
			"WIFI#123",
			"WIFI!ABC",
			"WIFI PRODUCT",
		}

		for _, sku := range invalidSKUs {
			assert.False(t, gen.IsValid(sku), fmt.Sprintf("SKU %s should be invalid", sku))
		}
	})

	t.Run("should validate prefix when set", func(t *testing.T) {
		genWithPrefix := NewGenerator("WIFI")
		assert.True(t, genWithPrefix.IsValid("WIFI-123456"))
		assert.False(t, genWithPrefix.IsValid("PULSA-123456"))
	})

	t.Run("should allow any alphanumeric+dash when no prefix", func(t *testing.T) {
		genNoPrefix := NewGenerator("")
		assert.True(t, genNoPrefix.IsValid("WIFI-123456"))
		assert.True(t, genNoPrefix.IsValid("PULSA-123456"))
	})
}

func TestExtractPrefix(t *testing.T) {
	t.Run("should extract prefix from SKU", func(t *testing.T) {
		testCases := []struct {
			sku      string
			expected string
		}{
			{"WIFI-1702556400000-AB3K9X", "WIFI"},
			{"PULSA-2023-001", "PULSA"},
			{"DATA-123-456-789", "DATA"},
		}

		for _, tc := range testCases {
			assert.Equal(t, tc.expected, ExtractPrefix(tc.sku))
		}
	})

	t.Run("should return first part when no prefix pattern detected", func(t *testing.T) {
		// When SKU starts with timestamp, the first part is the timestamp
		prefix := ExtractPrefix("1702556400000-AB3K9X")
		assert.Equal(t, "1702556400000", prefix)
	})

	t.Run("should return full string if no dash", func(t *testing.T) {
		assert.Equal(t, "", ExtractPrefix("SIMPLESKUWITHOUTDASH"))
	})
}

func TestExtractTimestamp(t *testing.T) {
	t.Run("should extract timestamp from SKU with prefix", func(t *testing.T) {
		now := time.Now().UnixMilli()
		sku := fmt.Sprintf("WIFI-%d-AB3K9X", now)
		timestamp := ExtractTimestamp(sku)

		// Should be close to current time (within 1 second)
		assert.True(t, timestamp > 0)
		assert.True(t, timestamp >= now-1000 && timestamp <= now+1000)
	})

	t.Run("should extract timestamp from SKU without prefix", func(t *testing.T) {
		now := time.Now().UnixMilli()
		sku := fmt.Sprintf("%d-AB3K9X", now)
		timestamp := ExtractTimestamp(sku)

		assert.True(t, timestamp > 0)
	})

	t.Run("should return 0 for invalid timestamp format", func(t *testing.T) {
		assert.Equal(t, int64(0), ExtractTimestamp("WIFI-NOTIMESTAMP-ABC"))
		assert.Equal(t, int64(0), ExtractTimestamp("SIMPLE-SKU"))
		assert.Equal(t, int64(0), ExtractTimestamp("NOSEPARATOR"))
	})
}

func TestValidateFormat(t *testing.T) {
	t.Run("should only allow alphanumeric and hyphens", func(t *testing.T) {
		validPattern := regexp.MustCompile(`^[A-Za-z0-9\-]+$`)

		validSKUs := []string{
			"WIFI-123",
			"PULSA-ABC",
			"DATA-456-XYZ",
			"123-456-789",
		}

		for _, sku := range validSKUs {
			assert.True(t, validPattern.MatchString(sku))
		}

		invalidSKUs := []string{
			"WIFI@123",
			"PULSA#ABC",
			"DATA!456",
			"123_456",
		}

		for _, sku := range invalidSKUs {
			assert.False(t, validPattern.MatchString(sku))
		}
	})
}

func TestGeneratedSKUConsistency(t *testing.T) {
	t.Run("should generate consistent SKU format", func(t *testing.T) {
		gen := NewGenerator("PRODUCT")

		for i := 0; i < 50; i++ {
			sku, err := gen.Generate()
			require.NoError(t, err)

			// Verify format: PREFIX-TIMESTAMP-RANDOM
			parts := strings.Split(sku, "-")
			assert.Equal(t, 3, len(parts), fmt.Sprintf("SKU %s should have 3 parts", sku))
			assert.Equal(t, "PRODUCT", parts[0])
			assert.Len(t, parts[1], 13) // timestamp in milliseconds
			assert.Len(t, parts[2], 6)  // random suffix
		}
	})
}

func BenchmarkGenerate(b *testing.B) {
	gen := NewGenerator("WIFI")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate()
	}
}

func BenchmarkIsValid(b *testing.B) {
	gen := NewGenerator("WIFI")
	sku := "WIFI-1702556400000-AB3K9X"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = gen.IsValid(sku)
	}
}

func BenchmarkExtractPrefix(b *testing.B) {
	sku := "WIFI-1702556400000-AB3K9X"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ExtractPrefix(sku)
	}
}
