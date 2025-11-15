# Example Schemas

This directory contains example schemas to help you get started with datagen. Each example demonstrates different features and use cases.

## Available Examples

### 1. Blog (`blog.json`)

A simple blog application with authors, posts, comments, categories, and tags.

**Features:**
- 6 tables with realistic relationships
- Nested comments (parent-child relationships)
- Many-to-many relationship (posts and tags)
- Weighted enum for post and comment status
- Pattern generators for slugs
- ~4,170 total rows

**Use case:** Content management systems, blogging platforms

**Generate:**
```bash
datagen generate --input docs/examples/blog.json --output blog.sql
```

**Or use COPY format for faster loading:**
```bash
datagen generate --input docs/examples/blog.json --output blog.sql --format copy
```

---

### 2. Analytics (`analytics.json`)

Web analytics tracking system with sessions, events, page views, and conversions.

**Features:**
- 5 tables with time-series data
- Event tracking with JSONB properties
- Device, browser, and OS tracking
- Timeseries generators for realistic temporal data
- Pattern generators for unique IDs
- Business hours pattern for conversions
- ~415,000 total rows (high volume)

**Use case:** Analytics platforms, event tracking, user behavior analysis

**Generate:**
```bash
datagen generate --input docs/examples/analytics.json --output analytics.sql --seed 42
```

**Tips:**
- Use `--seed` for reproducible datasets
- Consider using `--format copy` for this large dataset
- Adjust row_count values for testing vs. production scenarios

---

### 3. E-commerce Custom (`ecommerce-custom.json`)

Advanced multi-merchant e-commerce platform with complex order processing.

**Features:**
- 7 tables with marketplace features
- Multi-merchant support
- Advanced order management with multiple statuses
- Shipment tracking
- Loyalty points and lifetime value tracking
- Cost and profit margin data
- Template generators for order numbers
- ~50,100 total rows

**Use case:** Marketplaces, advanced e-commerce platforms, order management systems

**Generate:**
```bash
datagen generate --input docs/examples/ecommerce-custom.json --output ecommerce.sql --format sql
```

**Or validate first:**
```bash
datagen validate --input docs/examples/ecommerce-custom.json
```

---

## Customizing Examples

All examples can be customized by:

1. **Adjusting row counts:**
   ```json
   {
     "tables": {
       "users": {
         "row_count": 1000  // Change this value
       }
     }
   }
   ```

2. **Modifying weighted distributions:**
   ```json
   {
     "generator": "weighted_enum",
     "generator_config": {
       "values": ["active", "inactive"],
       "weights": [90, 10]  // Adjust weights
     }
   }
   ```

3. **Changing time ranges:**
   ```json
   {
     "generator": "timeseries",
     "generator_config": {
       "start": "2024-01-01T00:00:00Z",
       "end": "2024-12-31T23:59:59Z",
       "pattern": "uniform",
       "interval": "1h"
     }
   }
   ```

## Common Workflows

### Validate → Generate → Load

```bash
# 1. Validate schema
datagen validate --input docs/examples/blog.json

# 2. Generate SQL dump
datagen generate --input docs/examples/blog.json --output /tmp/blog.sql

# 3. Load into PostgreSQL
psql -d mydb -f /tmp/blog.sql
```

### Quick Test Data

```bash
# Generate with a specific seed for reproducible data
datagen generate --input docs/examples/analytics.json --output test.sql --seed 12345
```

### Large Datasets

```bash
# Use COPY format for faster loading of large datasets
datagen generate \
  --input docs/examples/analytics.json \
  --output analytics.sql \
  --format copy \
  --seed 42
```

## Generator Types Used

| Generator | Description | Example |
|-----------|-------------|---------|
| **semantic** | Auto-detects from column names | `email`, `first_name`, `phone` |
| **pattern** | Regex-based value generation | `"pattern": "[A-Z]{3}-[0-9]{8}"` |
| **template** | Template with placeholders | `"template": "ORD-{{year}}-{{seq:6}}"` |
| **weighted_enum** | Random selection with weights | `"values": ["active"], "weights": [80]` |
| **integer_range** | Random integers in range | `"min": 0, "max": 1000` |
| **timeseries** | Time-based sequential data | `"pattern": "uniform", "interval": "1h"` |

## Tips

1. **Start Small:** Begin with low row counts to test your schema before generating large datasets
2. **Use Seeds:** Use `--seed` flag for reproducible test data
3. **Validate First:** Always run `datagen validate` before generating large datasets
4. **Choose Format Wisely:**
   - Use `sql` format for small datasets or when you need readable INSERT statements
   - Use `copy` format for large datasets (much faster to load)
5. **Test Restore:** Always test that generated dumps can be restored to PostgreSQL

## Need Help?

- Run `datagen --help` for all available commands
- Run `datagen generate --help` for generation options
- Check pre-built templates with `datagen template list`
- See the full documentation at https://github.com/NhaLeTruc/datagen-cli

## Creating Your Own Schemas

Use these examples as templates for your own schemas. Key things to remember:

1. Always include `version` and `database` sections
2. Define tables with proper column types (see PostgreSQL documentation)
3. Set up foreign keys to maintain referential integrity
4. Use appropriate generators for realistic data
5. Test with small row counts first
6. Validate before generating large datasets

Happy data generating!
