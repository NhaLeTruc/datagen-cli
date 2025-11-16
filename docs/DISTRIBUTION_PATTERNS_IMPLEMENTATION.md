# Distribution & Pattern Implementation Guide

This document explains how distribution logic and custom patterns are implemented in datagen-cli.

## Overview

Three new advanced features extend the data generation capabilities:

1. **Distribution Logic** - Generate data following statistical distributions (weighted, normal, Poisson, Zipf)
2. **Custom Patterns** - Template-based generation with placeholders (`{year}`, `{sequence:6}`, etc.)
3. **Business Rules** - Conditional logic where one column's value depends on others

## Architecture

### Generator Hierarchy

```
Generator Interface (Generate(ctx) -> value)
    │
    ├── BasicGenerator (integer, varchar, timestamp)
    ├── SemanticGenerator (email, phone, name)
    │
    └── Advanced Generators (NEW):
        ├── DistributionGenerator
        ├── PatternGenerator
        └── RulesGenerator
```

### Selection Priority

When generating a column value, the system checks in this order:

1. **Business Rules** (highest priority) - Column has `rules: [...]`
2. **Distribution** - Column has `distribution: {...}`
3. **Pattern** - Column has `pattern: {...}`
4. **Custom Generator** - Column has `generator: "name"`
5. **Semantic Detection** - Auto-detect from column name
6. **Basic Type** (fallback) - Based on SQL type

## Files Created

### Core Implementation

1. **`internal/schema/distribution.go`**
   - Data structures for distribution configs
   - `DistributionConfig`, `BusinessRule`, `PatternConfig`

2. **`internal/generator/distribution.go`**
   - Distribution generator implementation
   - Supports: weighted, normal, Poisson, Zipf distributions

3. **`internal/generator/pattern.go`**
   - Pattern/template generator
   - Placeholder resolution: `{year}`, `{month}`, `{sequence:N}`, `{random:N}`, `{uuid}`, `{hex:N}`, `{alpha:N}`, `{alphanumeric:N}`, `{row}`, `{table}`, `{timestamp}`

4. **`internal/generator/rules.go`**
   - Business rules engine
   - Conditional logic: if-then-else
   - Cross-column dependencies

### Documentation & Examples

5. **`docs/examples/integration-example.go`**
   - Shows how pipeline integrates these generators

6. **`docs/examples/advanced-schema-examples.json`**
   - Real-world usage examples
   - E-commerce scenario with all features

7. **`docs/DISTRIBUTION_PATTERNS_IMPLEMENTATION.md`**
   - This file - complete implementation guide

## Schema Extensions

### Column Type (Extended)

```go
type Column struct {
    // ... existing fields ...

    // NEW: Advanced generation
    Distribution    *DistributionConfig    `json:"distribution,omitempty"`
    Pattern         *PatternConfig         `json:"pattern,omitempty"`
    Rules           []*BusinessRule        `json:"rules,omitempty"`
}
```

### Context Type (Extended)

```go
type Context struct {
    Rand       *rand.Rand
    Faker      *gofakeit.Faker  // NEW
    TableName  string
    ColumnName string
    RowNumber  int              // NEW
    RowData    map[string]interface{}  // NEW - for business rules
    data       map[string]interface{}
}
```

## Usage Examples

### 1. Weighted Distribution

**80% completed, 15% pending, 5% cancelled:**

```json
{
  "status": {
    "type": "varchar(20)",
    "distribution": {
      "type": "weighted",
      "weights": {
        "completed": 80,
        "pending": 15,
        "cancelled": 5
      }
    }
  }
}
```

### 2. Normal Distribution

**View counts normally distributed around 500:**

```json
{
  "view_count": {
    "type": "integer",
    "distribution": {
      "type": "normal",
      "mean": 500,
      "std_dev": 200,
      "min": 0,
      "max": 5000
    }
  }
}
```

### 3. Zipf Distribution (Power-Law)

**Few products are very popular (long tail):**

```json
{
  "popularity_score": {
    "type": "integer",
    "distribution": {
      "type": "zipf",
      "alpha": 1.5,
      "min": 1,
      "max": 1000
    }
  }
}
```

### 4. Poisson Distribution

**Events per hour (lambda=12.5):**

```json
{
  "hourly_events": {
    "type": "integer",
    "distribution": {
      "type": "poisson",
      "mean": 12.5
    }
  }
}
```

### 5. Custom Patterns

**Order numbers: `ORD-2024-00000123`:**

```json
{
  "order_number": {
    "type": "varchar(50)",
    "pattern": {
      "template": "ORD-{year}-{sequence:8}"
    }
  }
}
```

**API Keys: `pk_aB3dE5fG7hJ9kL2m`:**

```json
{
  "api_key": {
    "type": "varchar(64)",
    "pattern": {
      "template": "pk_{alphanumeric:32}"
    }
  }
}
```

**Event IDs: `evt_1634567890_uuid-here`:**

```json
{
  "event_id": {
    "type": "varchar(100)",
    "pattern": {
      "template": "evt_{timestamp}_{uuid}"
    }
  }
}
```

### 6. Business Rules (Conditional Logic)

**Price depends on order type:**

```json
{
  "order_type": {
    "type": "varchar(20)",
    "distribution": {
      "type": "weighted",
      "weights": {
        "standard": 70,
        "premium": 25,
        "enterprise": 5
      }
    }
  },
  "total_amount": {
    "type": "decimal(10,2)",
    "rules": [
      {
        "if": {"order_type": "standard"},
        "then": {"min": 10.00, "max": 99.99}
      },
      {
        "if": {"order_type": "premium"},
        "then": {"min": 100.00, "max": 999.99}
      },
      {
        "if": {"order_type": "enterprise"},
        "then": {"min": 1000.00, "max": 50000.00}
      }
    ]
  }
}
```

## Available Placeholders

Pattern templates support these placeholders:

| Placeholder | Description | Example |
|------------|-------------|---------|
| `{year}` | Current year | `2024` |
| `{month}` | Current month (number) | `03` |
| `{month:name}` | Month name | `March` |
| `{day}` | Current day | `15` |
| `{timestamp}` | Unix timestamp | `1634567890` |
| `{sequence}` | Auto-incrementing | `1`, `2`, `3` |
| `{sequence:N}` | Padded sequence | `00001`, `00002` |
| `{random:N}` | Random N-digit number | `847592` |
| `{uuid}` | UUIDv4 | `550e8400-e29b-...` |
| `{hex:N}` | Random hex (N chars) | `a3f5d2` |
| `{alpha:N}` | Random letters | `aBcDeF` |
| `{alphanumeric:N}` | Random alphanumeric | `aB3dE5` |
| `{row}` | Current row number | `42` |
| `{table}` | Current table name | `users` |

## Testing Strategy

### Unit Tests

1. **Distribution Generator Tests**
   ```go
   func TestWeightedDistribution(t *testing.T)
   func TestNormalDistribution(t *testing.T)
   func TestPoissonDistribution(t *testing.T)
   func TestZipfDistribution(t *testing.T)
   ```

2. **Pattern Generator Tests**
   ```go
   func TestSequencePlaceholder(t *testing.T)
   func TestRandomPlaceholder(t *testing.T)
   func TestDatePlaceholders(t *testing.T)
   func TestUUIDPlaceholder(t *testing.T)
   func TestComplexTemplate(t *testing.T)
   ```

3. **Rules Generator Tests**
   ```go
   func TestSimpleCondition(t *testing.T)
   func TestMultipleRules(t *testing.T)
   func TestRangeGeneration(t *testing.T)
   func TestFallbackBehavior(t *testing.T)
   ```

### Integration Tests

```go
func TestDistributionIntegration(t *testing.T) {
    // Generate 10,000 rows with 80/15/5 distribution
    // Verify distribution is within ±2% of expected
}

func TestBusinessRulesIntegration(t *testing.T) {
    // Generate rows where price depends on type
    // Verify all prices respect their type's range
}
```

### Statistical Validation

```go
func TestDistributionAccuracy(t *testing.T) {
    // For weighted: Chi-squared test
    // For normal: Kolmogorov-Smirnov test
    // For Poisson: Goodness-of-fit test
}
```

## Performance Considerations

1. **Sequence Tracking**
   - Sequences stored per-pattern in memory
   - Reset between tables if needed
   - Thread-safe with mutex

2. **Rule Evaluation**
   - Rules evaluated in order (short-circuit)
   - Row data cached in context
   - No database lookups during generation

3. **Distribution Sampling**
   - Pre-computed cumulative weights for weighted distribution
   - Standard library functions for statistical distributions
   - O(1) sampling for most distributions

## Migration Path

### Phase 1: Core Implementation
1. Implement distribution.go ✓
2. Implement pattern.go ✓
3. Implement rules.go ✓
4. Update schema types ✓

### Phase 2: Integration
1. Update Context with RowData
2. Modify pipeline to track row state
3. Update generator selection logic
4. Register new generators in coordinator

### Phase 3: Testing
1. Unit tests for each generator
2. Integration tests with real schemas
3. Statistical validation tests
4. Performance benchmarks

### Phase 4: Documentation
1. Update schema format documentation
2. Add examples to quickstart
3. Update README with examples
4. Add to pre-built templates

## Next Steps

To activate these features:

1. **Update Column Type** in `internal/schema/types.go`
   - Add Distribution, Pattern, Rules fields

2. **Update Context** in `internal/generator/context.go`
   - Add Faker, RowNumber, RowData fields

3. **Update Pipeline** in `internal/pipeline/coordinator.go`
   - Modify row generation to track RowData
   - Update generator selection logic

4. **Register Generators** in initialization
   ```go
   func init() {
       // These would be registered when column config is parsed
       // Not registered globally since they're config-driven
   }
   ```

5. **Add Tests**
   - Create test files for each new generator
   - Add integration test with complete schema

6. **Update Templates**
   - Add distribution examples to ecommerce template
   - Add pattern examples to SaaS template

## Example: Complete Flow

```
User Schema:
  orders.status has weighted distribution

Pipeline Processing:
  1. Parse schema → detect distribution config
  2. Create DistributionGenerator(config)
  3. For each row:
     - Call generator.Generate(ctx)
     - Weighted sampling returns "completed" (80% chance)
     - Store in row["status"]
  4. Write to SQL dump

Result:
  10,000 rows → ~8,000 "completed", ~1,500 "pending", ~500 "cancelled"
```

## FAQ

**Q: Can I combine distribution with business rules?**
A: Yes! Rules have highest priority and can reference distributed columns.

**Q: Are sequences deterministic with seeds?**
A: Yes, with the same seed, sequences will be identical across runs.

**Q: Can patterns reference other columns?**
A: Currently no, but you can use business rules to achieve similar results.

**Q: What happens if weights don't sum to 100?**
A: Weights are normalized automatically (they can be any positive numbers).

**Q: Can I use normal distribution with strings?**
A: No, normal/Poisson/Zipf are for numeric types only. Use weighted for strings.

## References

- [User Story 3 Specification](../specs/001-json-to-pgdump/spec.md)
- [Statistical Distributions](https://en.wikipedia.org/wiki/List_of_probability_distributions)
- [Zipf's Law](https://en.wikipedia.org/wiki/Zipf%27s_law)
- [Pattern Syntax Design](./examples/advanced-schema-examples.json)
