# Quickstart Guide: datagen-cli

**Feature**: 001-json-to-pgdump
**Version**: 1.0
**Created**: 2025-11-15

## What is datagen-cli?

datagen-cli is a command-line tool that transforms declarative JSON schema definitions into fully-formed PostgreSQL dump files containing realistic mock data, without requiring a running PostgreSQL instance.

**Perfect for**:
- Creating realistic test databases for local development
- Generating demo data for presentations
- CI/CD pipelines needing fresh test data
- Compliance-safe testing without production data

## Installation

### Option 1: Go Install (requires Go 1.21+)
```bash
go install github.com/NhaLeTruc/datagen-cli/cmd/datagen@latest
```

### Option 2: Homebrew (macOS/Linux)
```bash
brew tap NhaLeTruc/datagen
brew install datagen
```

### Option 3: Download Binary
Download from [GitHub Releases](https://github.com/NhaLeTruc/datagen-cli/releases) for your platform.

### Option 4: Build from Source
```bash
git clone https://github.com/NhaLeTruc/datagen-cli.git
cd datagen-cli
go build -o datagen ./cmd/datagen
```

### Verify Installation
```bash
datagen version
```

## 5-Minute Tutorial

### Step 1: Create a Simple Schema

Create `blog-schema.json`:
```json
{
  "version": "1.0",
  "database": {
    "name": "blog",
    "encoding": "UTF8",
    "locale": "en_US.utf8"
  },
  "tables": {
    "users": {
      "columns": {
        "id": {
          "type": "serial",
          "primary_key": true
        },
        "email": {
          "type": "varchar(255)",
          "nullable": false,
          "unique": true
        },
        "username": {
          "type": "varchar(50)",
          "nullable": false
        },
        "created_at": {
          "type": "timestamp",
          "nullable": false,
          "default": "now()"
        }
      },
      "row_count": 100
    },
    "posts": {
      "columns": {
        "id": {
          "type": "serial",
          "primary_key": true
        },
        "user_id": {
          "type": "integer",
          "nullable": false
        },
        "title": {
          "type": "varchar(200)",
          "nullable": false
        },
        "content": {
          "type": "text",
          "nullable": false,
          "generator": "lorem_ipsum",
          "generator_config": {
            "paragraphs": 3
          }
        },
        "published_at": {
          "type": "timestamp",
          "nullable": false
        }
      },
      "foreign_keys": [
        {
          "columns": ["user_id"],
          "referenced_table": "users",
          "referenced_columns": ["id"],
          "on_delete": "CASCADE"
        }
      ],
      "indexes": [
        {"columns": ["user_id"]},
        {"columns": ["published_at"]}
      ],
      "row_count": 500
    }
  }
}
```

### Step 2: Generate the Dump

```bash
datagen generate blog-schema.json -o blog.dump --verbose
```

Expected output:
```
Parsing schema... ✓
Validating schema... ✓
Resolving dependencies... ✓
  Found 2 tables in dependency order: users, posts

Generating data...
  ├─ users [100/100 rows] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% (0.2s)
  └─ posts [500/500 rows] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% (1.1s)

Writing dump file... ✓
Finalizing... ✓

Generated 600 rows across 2 tables in 1.3s
Output: blog.dump (124 KB compressed)
```

### Step 3: Restore to PostgreSQL

```bash
# Create database
createdb blog

# Restore dump
pg_restore -d blog blog.dump

# Verify
psql -d blog -c "SELECT COUNT(*) FROM users;"
psql -d blog -c "SELECT COUNT(*) FROM posts;"
```

🎉 **Congratulations!** You just generated a realistic test database in seconds.

## Using Templates

datagen includes pre-built templates for common scenarios.

### Available Templates
```bash
datagen template --list
```

Output:
```
Available Templates:

  ecommerce
    E-commerce platform with products, orders, customers
    Tables: products, categories, customers, orders, order_items, reviews

  saas
    SaaS application with tenants, users, subscriptions
    Tables: tenants, users, subscriptions, usage_metrics, billing

  healthcare
    Healthcare system with patients, appointments, medical records
    Tables: patients, doctors, appointments, medical_records, prescriptions

  finance
    Financial system with accounts, transactions, investments
    Tables: accounts, customers, transactions, investments, portfolios
```

### Generate from Template
```bash
# Use ecommerce template with 10,000 rows per table
datagen generate --template ecommerce --template-param rows=10000 -o ecommerce.dump

# Customize specific parameters
datagen generate --template ecommerce \
  --template-param customers=5000 \
  --template-param products=2000 \
  --template-param orders=15000 \
  -o ecommerce-large.dump
```

### Export Template for Customization
```bash
# Export template to JSON
datagen template ecommerce --export ecommerce-base.json

# Edit the JSON file
vim ecommerce-base.json

# Generate with your customizations
datagen generate ecommerce-base.json -o custom.dump
```

## Advanced Features

### Deterministic Data Generation

Generate identical data across runs using seeds:

```bash
# Generate with seed
datagen generate schema.json --seed 12345 -o data-v1.dump

# Later, regenerate identical data
datagen generate schema.json --seed 12345 -o data-v2.dump

# Verify they're identical
diff data-v1.dump data-v2.dump  # No differences
```

Perfect for:
- Reproducible test environments
- CI/CD pipelines
- Debugging specific data scenarios

### Different Output Formats

```bash
# Custom format (default, binary, compressed)
datagen generate schema.json -o output.dump

# SQL format (human-readable INSERT statements)
datagen generate schema.json --format sql -o output.sql
psql -d mydb -f output.sql

# COPY format (fastest import)
datagen generate schema.json --format copy -o output.copy.sql
psql -d mydb -f output.copy.sql
```

### Schema Validation

Validate schema before generating data:

```bash
# Quick validation
datagen validate schema.json

# Detailed validation with suggestions
datagen validate schema.json --verbose

# JSON output for automation
datagen validate schema.json --json
```

### Custom Data Generators

Control exactly how data is generated:

```json
{
  "tables": {
    "orders": {
      "columns": {
        "status": {
          "type": "varchar(20)",
          "generator": "weighted_enum",
          "generator_config": {
            "values": {
              "completed": 0.80,
              "pending": 0.15,
              "cancelled": 0.05
            }
          }
        },
        "order_number": {
          "type": "varchar(50)",
          "generator": "template",
          "generator_config": {
            "template": "ORD-{{year}}-{{seq:6}}"
          }
        },
        "amount": {
          "type": "numeric(10,2)",
          "generator": "integer_range",
          "generator_config": {
            "min": 10,
            "max": 1000
          }
        }
      },
      "row_count": 10000
    }
  }
}
```

### Time-Series Data

Generate realistic time-series data with patterns:

```json
{
  "columns": {
    "created_at": {
      "type": "timestamp",
      "generator": "timeseries",
      "generator_config": {
        "start": "2024-01-01T00:00:00Z",
        "end": "2024-12-31T23:59:59Z",
        "interval": "1h",
        "pattern": "business_hours"
      }
    }
  }
}
```

Patterns:
- `uniform`: Evenly distributed
- `business_hours`: 9 AM - 5 PM weekdays
- `daily_peak`: Higher activity during peak hours
- `weekly_pattern`: Weekend vs weekday distribution

## Common Workflows

### Development Workflow

```bash
# 1. Create schema for your app
vim app-schema.json

# 2. Validate schema
datagen validate app-schema.json

# 3. Generate small test dataset
datagen generate app-schema.json --rows 100 -o dev.dump

# 4. Restore to local database
dropdb --if-exists myapp_dev && createdb myapp_dev
pg_restore -d myapp_dev dev.dump

# 5. Run your app
npm run dev
```

### CI/CD Pipeline

```bash
#!/bin/bash
# Generate fresh test data for each CI run

# Use CI build number as seed for reproducibility
BUILD_NUM=${CI_BUILD_NUMBER:-1}

datagen generate test-schema.json \
  --seed $BUILD_NUM \
  --rows 1000 \
  -o test-data.dump

# Restore to test database
pg_restore -d test_db test-data.dump

# Run tests
npm test
```

### Creating Demo Data

```bash
# Generate compelling demo with realistic scale
datagen generate --template saas \
  --template-param customers=10000 \
  --template-param tenants=500 \
  --template-param subscriptions=8000 \
  --seed 42 \
  -o demo.dump \
  --verbose

# Restore for demo
pg_restore -d demo_db demo.dump
```

## Configuration File

Create `.datagen.yaml` in your project or home directory:

```yaml
# Default settings
format: custom
compress: true
rows: 1000
jobs: 4
verbose: false

# Template customizations
templates:
  ecommerce:
    rows: 5000
    customers: 1000
    products: 2000

  saas:
    rows: 3000
    tenants: 200
```

Settings precedence (highest to lowest):
1. CLI flags
2. `--config` file
3. `.datagen.yaml` in current directory
4. `~/.datagen.yaml`
5. Defaults

## Troubleshooting

### Schema Validation Errors

**Error**: `Unknown column type "varcharr"`
```
Fix: Check for typos in column types. Use "varchar(N)" not "varcharr".
```

**Error**: `Circular dependency detected: orders -> order_items -> orders`
```
Fix: Break circular foreign key dependencies. Use nullable foreign keys or remove one FK.
```

**Error**: `Foreign key references non-existent table "user"`
```
Fix: Ensure table name matches exactly. PostgreSQL is case-sensitive in quotes.
```

### Generation Errors

**Error**: `Out of memory`
```
Solution: Reduce row counts or enable streaming with --jobs 1
```

**Error**: `Failed to generate unique values`
```
Solution: Increase row count or remove unique constraint.
For unique constraints, ensure row_count < generator's unique value space.
```

### Restoration Errors

**Error**: `pg_restore: error: could not execute query`
```
Solution: Validate dump with --validate-output flag during generation.
Check PostgreSQL version compatibility (12-16 supported).
```

## Getting Help

```bash
# General help
datagen --help

# Command-specific help
datagen generate --help
datagen validate --help
datagen template --help

# Show version
datagen version
```

## Next Steps

- **Read [Schema Format Guide](contracts/schema-format.md)** - Learn all generator options
- **Read [CLI Reference](contracts/cli.md)** - Complete CLI documentation
- **Browse Examples** - See `docs/examples/` for real-world schemas
- **Join Community** - GitHub Discussions for questions and sharing

## Examples Repository

Check `docs/examples/` for complete working examples:
- `blog/` - Simple blog schema
- `ecommerce/` - Full e-commerce platform
- `saas/` - Multi-tenant SaaS application
- `analytics/` - Time-series analytics data
- `social-network/` - Social network with users, posts, follows

## FAQ

**Q: Can I use datagen with other databases (MySQL, MongoDB)?**
A: Currently datagen only supports PostgreSQL. Support for other databases may come in future versions.

**Q: Is the generated data GDPR compliant?**
A: Yes! All data is synthetic and fake. No real personal information is used.

**Q: Can I use this for performance testing?**
A: Absolutely! Generate large datasets with realistic distributions to test query performance.

**Q: How do I generate data matching specific business rules?**
A: Use custom generators with weighted enums, ranges, and patterns. See schema-format.md for details.

**Q: Can I extend datagen with my own generators?**
A: Yes! datagen supports custom generator plugins (coming in v1.1). For now, use pattern and template generators.

**Q: What's the largest dataset datagen can create?**
A: Tested up to 100GB dumps. Memory usage stays under 500MB thanks to streaming architecture.

## Performance Tips

1. **Use Custom Format**: 3-5x smaller than SQL, faster to restore
2. **Enable Compression**: Enabled by default for custom format
3. **Parallel Workers**: Use `--jobs N` to parallelize generation
4. **COPY Format for Speed**: Fastest import, but less flexible
5. **Deterministic Seeds**: Reuse generated data instead of regenerating

## Contributing

Found a bug? Want a feature?
- Open an issue: https://github.com/NhaLeTruc/datagen-cli/issues
- Submit a PR: https://github.com/NhaLeTruc/datagen-cli/pulls

## License

MIT License - see LICENSE file for details