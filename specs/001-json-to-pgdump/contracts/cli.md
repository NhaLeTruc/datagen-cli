# CLI Interface Contract

**Feature**: 001-json-to-pgdump
**Version**: 1.0
**Created**: 2025-11-15

## Command Structure

```
datagen <command> [arguments] [flags]
```

## Commands

### `generate`

Generate PostgreSQL dump file from JSON schema definition.

**Usage**:
```bash
datagen generate <schema-file> [flags]
datagen generate --template <template-name> [flags]
```

**Arguments**:
- `<schema-file>`: Path to JSON schema file (use `-` for stdin)

**Flags**:
- `--output, -o <file>`: Output file path (default: stdout)
- `--format <format>`: Output format: `custom`, `sql`, `copy` (default: `custom`)
- `--seed <value>`: Random seed for deterministic output (integer)
- `--rows <count>`: Default row count per table (overrides schema) (default: 1000)
- `--template <name>`: Use pre-built template (`ecommerce`, `saas`, `healthcare`, `finance`)
- `--template-param <key=value>`: Set template parameter (repeatable)
- `--config <file>`: Configuration file path (YAML or JSON)
- `--verbose, -v`: Verbose output with progress details
- `--dry-run`: Validate schema only, don't generate output
- `--validate-output`: Validate generated dump with pg_query (slower but safer)
- `--compress`: Enable gzip compression (only for custom format) (default: true)
- `--jobs <count>`: Number of parallel workers (default: CPU cores, max: 16)

**Examples**:
```bash
# Basic usage with schema file
datagen generate schema.json -o output.dump

# Use template with customization
datagen generate --template ecommerce --template-param rows=10000 -o demo.dump

# Generate SQL instead of custom format
datagen generate schema.json --format sql -o output.sql

# Deterministic generation with seed
datagen generate schema.json --seed 12345 -o reproducible.dump

# Read from stdin, write to stdout
cat schema.json | datagen generate - > output.dump

# Dry-run to validate schema
datagen generate schema.json --dry-run

# Verbose mode with validation
datagen generate schema.json -o out.dump -v --validate-output
```

**Exit Codes**:
- `0`: Success
- `1`: General error (invalid arguments, file not found)
- `2`: Schema validation error
- `3`: Data generation error
- `4`: Output write error
- `5`: Validation error (when --validate-output is used)

**Output** (stdout when -o not specified):
- Binary dump data (custom format) or SQL text (sql format)

**Output** (stderr):
- Progress indicators (when --verbose)
- Error messages
- Warnings

---

### `validate`

Validate JSON schema without generating data.

**Usage**:
```bash
datagen validate <schema-file> [flags]
```

**Arguments**:
- `<schema-file>`: Path to JSON schema file (use `-` for stdin)

**Flags**:
- `--verbose, -v`: Show detailed validation results
- `--json`: Output validation results as JSON

**Examples**:
```bash
# Validate schema
datagen validate schema.json

# Validate with details
datagen validate schema.json --verbose

# JSON output for automation
datagen validate schema.json --json
```

**Exit Codes**:
- `0`: Schema is valid
- `2`: Schema validation failed

**Output** (stdout):
```
✓ Schema is valid
  - 5 tables, 23 columns
  - No circular dependencies
  - All foreign key references resolved
```

**Output** (stderr):
- Validation errors with line/column information

**JSON Output** (with --json):
```json
{
  "valid": true,
  "stats": {
    "tables": 5,
    "columns": 23,
    "foreign_keys": 4,
    "indexes": 7
  },
  "warnings": [],
  "errors": []
}
```

---

### `template`

List available templates or show template details.

**Usage**:
```bash
datagen template [template-name] [flags]
```

**Arguments**:
- `[template-name]`: Optional template name to show details

**Flags**:
- `--list, -l`: List all available templates (default if no name provided)
- `--json`: Output as JSON
- `--export <file>`: Export template schema to file

**Examples**:
```bash
# List all templates
datagen template

# Show template details
datagen template ecommerce

# Export template schema
datagen template ecommerce --export ecommerce-base.json

# JSON output
datagen template --list --json
```

**Output** (list):
```
Available Templates:

  ecommerce
    E-commerce platform with products, orders, customers
    Tables: products, categories, customers, orders, order_items, reviews
    Default rows: 1000

  saas
    SaaS application with tenants, users, subscriptions
    Tables: tenants, users, subscriptions, usage_metrics, billing
    Default rows: 500

  healthcare
    Healthcare system with patients, appointments, medical records
    Tables: patients, doctors, appointments, medical_records, prescriptions
    Default rows: 1000

  finance
    Financial system with accounts, transactions, investments
    Tables: accounts, customers, transactions, investments, portfolios
    Default rows: 2000
```

**Output** (details):
```
Template: ecommerce
Description: E-commerce platform with products, orders, customers
Category: retail

Tables:
  - products (1000 rows)
  - categories (50 rows)
  - customers (500 rows)
  - orders (2000 rows)
  - order_items (5000 rows)
  - reviews (3000 rows)

Parameters:
  - rows: Default row count (default: 1000)
  - customers: Number of customers (default: 500)
  - products: Number of products (default: 1000)

Usage:
  datagen generate --template ecommerce --template-param rows=5000
```

---

### `version`

Show version information.

**Usage**:
```bash
datagen version [flags]
```

**Flags**:
- `--short`: Show version number only
- `--json`: Output as JSON

**Examples**:
```bash
# Full version info
datagen version

# Short version
datagen version --short

# JSON output
datagen version --json
```

**Output** (full):
```
datagen version 1.0.0
Built: 2025-11-15T10:30:00Z
Go version: go1.21.5
Platform: linux/amd64
Commit: abc1234
```

**Output** (short):
```
1.0.0
```

**JSON Output**:
```json
{
  "version": "1.0.0",
  "build_time": "2025-11-15T10:30:00Z",
  "go_version": "go1.21.5",
  "platform": "linux/amd64",
  "commit": "abc1234"
}
```

---

## Global Flags

Available for all commands:

- `--help, -h`: Show help for command
- `--version`: Show version information (same as `datagen version`)
- `--config <file>`: Global configuration file

## Configuration File

Configuration can be provided via YAML or JSON file.

**Location Precedence**:
1. `--config` flag value
2. `.datagen.yaml` in current directory
3. `.datagen.yaml` in home directory (`~/.datagen.yaml`)

**Example** (`.datagen.yaml`):
```yaml
# Default output format
format: custom

# Default row count
rows: 1000

# Compression enabled
compress: true

# Number of parallel workers
jobs: 4

# Seed for deterministic generation (0 = random)
seed: 0

# Validation enabled
validate_output: false

# Verbose mode
verbose: false

# Template parameters
templates:
  ecommerce:
    rows: 5000
    customers: 1000
    products: 2000
```

**Example** (`.datagen.json`):
```json
{
  "format": "custom",
  "rows": 1000,
  "compress": true,
  "jobs": 4,
  "seed": 0,
  "validate_output": false,
  "verbose": false,
  "templates": {
    "ecommerce": {
      "rows": 5000,
      "customers": 1000,
      "products": 2000
    }
  }
}
```

## Environment Variables

- `DATAGEN_CONFIG`: Path to configuration file (overridden by --config flag)
- `DATAGEN_VERBOSE`: Enable verbose mode (0 or 1)
- `DATAGEN_SEED`: Default seed value
- `DATAGEN_JOBS`: Number of parallel workers

## Progress Output

When `--verbose` is enabled, progress information is written to stderr:

```
Parsing schema... ✓
Validating schema... ✓
Resolving dependencies... ✓
  Found 5 tables in dependency order: categories, products, customers, orders, order_items

Generating data...
  ├─ categories [50/50 rows] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% (0.1s)
  ├─ products [1000/1000 rows] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% (2.3s)
  ├─ customers [500/500 rows] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% (1.1s)
  ├─ orders [2000/2000 rows] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% (4.2s)
  └─ order_items [5000/5000 rows] ━━━━━━━━━━━━━━━━━━━━━━━━ 100% (10.5s)

Writing dump file... ✓
Finalizing... ✓

Generated 8550 rows across 5 tables in 18.2s
Output: output.dump (2.3 MB compressed)
```

## Error Messages

Errors are written to stderr with context and suggestions:

```
Error: Schema validation failed

  Line 15, column 8: Unknown column type "varcharr"
    Did you mean "varchar"?

  Line 23, column 12: Foreign key references non-existent table "user"
    Available tables: users, products, orders

  Line 31, column 5: Circular dependency detected: orders -> order_items -> orders

Fix these errors and try again.
```

## Exit Codes Summary

| Code | Meaning | Example |
|------|---------|---------|
| 0 | Success | Command completed successfully |
| 1 | General error | Invalid arguments, file not found |
| 2 | Schema validation error | Invalid JSON schema, constraint violations |
| 3 | Data generation error | Generator failure, out of memory |
| 4 | Output write error | Permission denied, disk full |
| 5 | Validation error | Generated dump failed pg_query validation |

## Compatibility

- **Stdin/Stdout**: Fully supports piping for scripting
- **POSIX Compliance**: Follows standard CLI conventions
- **Shell Integration**: Works with bash, zsh, fish completions (generated separately)
- **CI/CD**: Exit codes and JSON output enable automation

## Examples for Common Scenarios

### Development Workflow
```bash
# Quick test database
datagen generate --template saas --template-param rows=100 -o dev.dump
pg_restore -d myapp_dev dev.dump
```

### CI/CD Pipeline
```bash
# Generate reproducible test data
datagen generate test-schema.json --seed $CI_BUILD_NUMBER -o test.dump
if [ $? -eq 0 ]; then
  pg_restore -d test_db test.dump
  npm test
fi
```

### Demo Data
```bash
# Create compelling demo with large dataset
datagen generate --template ecommerce \
  --template-param customers=10000 \
  --template-param products=5000 \
  --seed 42 \
  -o demo.dump \
  --verbose
```

### Schema Validation in Pre-commit Hook
```bash
#!/bin/bash
datagen validate schemas/*.json --json | jq '.valid'
```