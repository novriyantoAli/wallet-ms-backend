# SKU Utility Package

Best practice SKU (Stock Keeping Unit) generation and validation utility for the wallet microservice.

## Features

- **Automatic SKU Generation**: Generate unique SKUs with timestamp and random suffix
- **Custom Patterns**: Support for custom SKU patterns with template variables
- **Prefix Support**: Optional prefix support for categorizing SKUs
- **Validation**: Comprehensive SKU format validation
- **Extraction**: Extract prefix and timestamp from generated SKUs

## Usage

### Basic Generation

```go
import "github.com/novriyantoAli/wallet-ms-backend/internal/pkg/sku"

// Create generator with prefix
gen := sku.NewGenerator("WIFI")
skuValue, err := gen.Generate()
// Output: WIFI-1702556400000-AB3K9X

// Create generator without prefix
gen := sku.NewGenerator("")
skuValue, err := gen.Generate()
// Output: 1702556400000-AB3K9X
```

### Custom Patterns

```go
// Using timestamp and random
gen := sku.NewGenerator("PULSA")
sku, _ := gen.GenerateWithPattern("PULSA-{timestamp}-{random}")
// Output: PULSA-1702556400000-XY7Z2M

// Using date only
sku, _ := gen.GenerateWithPattern("DATA-{date}")
// Output: DATA-20231214

// Complex pattern
sku, _ := gen.GenerateWithPattern("PRODUCT-{date}-{time}-{random}")
// Output: PRODUCT-20231214-153045-AB3K9X
```

### Validation

```go
gen := sku.NewGenerator("WIFI")

// Validate SKU format
valid := gen.IsValid("WIFI-1702556400000-AB3K9X")
// Returns: true

valid = gen.IsValid("INVALID@SKU")
// Returns: false
```

### Extraction

```go
// Extract prefix from SKU
prefix := sku.ExtractPrefix("WIFI-1702556400000-AB3K9X")
// Returns: "WIFI"

// Extract timestamp from generated SKU
timestamp := sku.ExtractTimestamp("WIFI-1702556400000-AB3K9X")
// Returns: 1702556400000 (Unix milliseconds)
```

## SKU Format

### Default Format
```
[PREFIX-]TIMESTAMP-RANDOM
```

- **PREFIX**: Optional category prefix (e.g., WIFI, PULSA, DATA)
- **TIMESTAMP**: Unix milliseconds (13 digits)
- **RANDOM**: Alphanumeric random suffix (6 characters)

### Pattern Variables

Supported template variables for custom patterns:

- `{prefix}`: Generator prefix
- `{timestamp}`: Unix milliseconds
- `{date}`: Date in YYYYMMDD format
- `{time}`: Time in HHmmss format
- `{random}`: 6-character random alphanumeric suffix

## Validation Rules

- Maximum length: 100 characters (database constraint)
- Allowed characters: A-Z, a-z, 0-9, hyphen (-)
- Required: Cannot be empty
- Prefix-based validation: If prefix is set, SKU must start with that prefix

## Best Practices

1. **Use Prefixes for Categories**
   ```go
   wifiGen := sku.NewGenerator("WIFI")
   pulsaGen := sku.NewGenerator("PULSA")
   ```

2. **Validate Before Database Operations**
   ```go
   if !gen.IsValid(skuValue) {
       return fmt.Errorf("invalid SKU format")
   }
   ```

3. **Extract Metadata When Needed**
   ```go
   timestamp := sku.ExtractTimestamp(skuValue)
   prefix := sku.ExtractPrefix(skuValue)
   ```

4. **Use Appropriate Pattern**
   - Time-based uniqueness: Use default `Generate()`
   - Date grouping: Use `{date}` pattern
   - Custom sequences: Define custom patterns

## Integration Example

```go
// In Product Service
type ProductService struct {
    skuGen *sku.Generator
}

func NewProductService() *ProductService {
    return &ProductService{
        skuGen: sku.NewGenerator("PRODUCT"),
    }
}

func (s *ProductService) CreateProduct(req *CreateProductRequest) error {
    // Generate unique SKU
    skuValue, err := s.skuGen.Generate()
    if err != nil {
        return err
    }

    product := &entity.Product{
        Name: req.Name,
        SKU:  skuValue,
        // ... other fields
    }

    return s.repo.Create(product)
}
```

## Testing

Run tests:
```bash
go test ./internal/pkg/sku -v
```

Run benchmarks:
```bash
go test ./internal/pkg/sku -bench=. -benchmem
```

## Thread Safety

The `Generator` type is thread-safe and can be safely used from multiple goroutines.

## Performance

- **Generation**: ~5-10 microseconds per SKU
- **Validation**: ~1-2 microseconds per check
- **Extraction**: ~500 nanoseconds per operation

See benchmarks in `sku_test.go` for detailed metrics.
