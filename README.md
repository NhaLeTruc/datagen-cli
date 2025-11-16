# datagen-cli

> Transform JSON schemas into PostgreSQL databases filled with realistic mock data

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/coverage-97%25-brightgreen.svg)](tests/)

**datagen-cli** is a command-line tool that transforms declarative JSON schema definitions into fully-formed PostgreSQL dump files containing realistic mock data—**no PostgreSQL instance required**.

Perfect for:
- 🧪 Creating realistic test databases for local development
- 🎯 Generating demo data for presentations and prototypes
- 🔄 CI/CD pipelines needing fresh, deterministic test data
- 🔒 Compliance-safe testing without using production data
- 📊 Populating databases for performance testing and benchmarking

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Configuration](#configuration)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

---

## Features

### Core Capabilities

- **🎯 Smart Data Generation**: Automatically generates realistic data based on column names
  - Email addresses, phone numbers, names, addresses
  - Timestamps, UUIDs, JSON, IP addresses
  - 200+ semantic patterns recognized

- **🔗 Referential Integrity**: Maintains foreign key relationships and proper data insertion order

- **📦 Multiple Output Formats**:
  - SQL format with INSERT statements
  - COPY format for faster loading
  - PostgreSQL custom dump format (planned)

- **✅ Built-in Validation**:
  - Schema validation before generation
  - Optional SQL syntax validation using PostgreSQL's actual parser
  - Detailed error messages with line numbers

- **🎲 Deterministic Generation**: Use seeds for reproducible test data

- **📚 Pre-built Templates**: Start quickly with templates for common scenarios:
  - E-commerce (products, orders, customers, reviews)
  - SaaS (tenants, users, subscriptions, billing)
  - Healthcare (patients, appointments, prescriptions)
  - Finance (accounts, transactions, ledgers)

- **⚡ Performance Optimized**:
  - Streaming architecture for large datasets
  - Worker pool for parallel processing
  - LRU cache for efficient foreign key lookups
  - Handles 100GB+ datasets

- **🔧 Developer-Friendly**:
  - Progress indicators and statistics
  - Structured logging with security event tracking
  - Configurable via files, environment variables, or CLI flags
  - Verbose mode for debugging

---

## Installation

### Option 1: Go Install (Requires Go 1.21+)

```bash
go install github.com/NhaLeTruc/datagen-cli/cmd/datagen@latest
```

### Option 2: Homebrew (macOS/Linux)

```bash
brew tap NhaLeTruc/datagen
brew install datagen
```

### Option 3: Download Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/NhaLeTruc/datagen-cli/releases).

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

---

## Quick Start

### 1. Create a Schema File

Create `blog.json`:

```json
{
  "version": "1.0",
  "database": {
    "name": "blog",
    "encoding": "UTF8"
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
          "nullable": false
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
          "nullable": false
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
          "referenced_columns": ["id"]
        }
      ],
      "row_count": 500
    }
  }
}
```

### 2. Generate SQL Dump

```bash
datagen generate -i blog.json -o blog.sql
```

### 3. Import to PostgreSQL

```bash
psql -U postgres -d blog -f blog.sql
```

---

## Usage

### Commands

#### `generate` - Generate mock data

```bash
# Basic usage
datagen generate -i schema.json -o dump.sql

# Use a pre-built template
datagen generate --template ecommerce -o ecommerce.sql

# Generate with deterministic seed
datagen generate -i schema.json -o dump.sql --seed 12345

# Use COPY format for faster loading
datagen generate -i schema.json -o dump.sql --format copy

# Validate SQL output
datagen generate -i schema.json -o dump.sql --validate-output

# Control parallel workers
datagen generate -i schema.json -o dump.sql --jobs 8

# Verbose output
datagen generate -i schema.json -o dump.sql --verbose
```

#### `validate` - Validate schema

```bash
# Validate a schema file
datagen validate -i schema.json

# Validate with detailed output
datagen validate -i schema.json --verbose

# Output validation results as JSON
datagen validate -i schema.json --json
```

#### `template` - Work with templates

```bash
# List available templates
datagen template list

# Show template details
datagen template show ecommerce

# Generate from template with custom parameters
datagen generate --template saas --param tenants=500 -o saas.sql
```

#### `version` - Show version information

```bash
datagen version
```

### Global Flags

- `--config` - Config file path (default: `.datagen.yaml`)
- `--verbose` - Enable verbose logging
- `--help` - Show help information

### Output Formats

**SQL Format (INSERT statements)**:
```sql
INSERT INTO users (id, email, username, created_at) VALUES
  (1, 'alice@example.com', 'alice', '2024-01-15 10:30:00'),
  (2, 'bob@example.com', 'bob', '2024-01-15 10:31:00');
```

**COPY Format (faster for large datasets)**:
```sql
COPY users (id, email, username, created_at) FROM stdin;
1	alice@example.com	alice	2024-01-15 10:30:00
2	bob@example.com	bob	2024-01-15 10:31:00
\.
```

---

## Architecture

datagen-cli uses a **pipeline architecture** designed for performance, extensibility, and PostgreSQL compatibility.

```
┌─────────────────────────────────────────────────────────────────┐
│                        datagen-cli                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
        ┌──────────────────────────────────────────┐
        │         CLI Layer (Cobra)                │
        │  • Command parsing                       │
        │  • Flag validation                       │
        │  • Progress reporting                    │
        └──────────────────────────────────────────┘
                              │
                              ▼
        ┌──────────────────────────────────────────┐
        │      Schema Layer                        │
        │  • JSON parsing                          │
        │  • Schema validation                     │
        │  • Type checking                         │
        └──────────────────────────────────────────┘
                              │
                              ▼
        ┌──────────────────────────────────────────┐
        │      Pipeline Coordinator                │
        │  • Dependency resolution                 │
        │  • Execution order                       │
        │  • Worker orchestration                  │
        └──────────────────────────────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
                    ▼                   ▼
        ┌──────────────────┐  ┌──────────────────┐
        │  Generator       │  │  Worker Pool     │
        │  Registry        │  │  • Parallel      │
        │  • Semantic      │  │    processing    │
        │    detection     │  │  • Backpressure  │
        │  • Custom        │  │  • Stats         │
        │    patterns      │  │                  │
        │  • FK cache      │  │                  │
        └──────────────────┘  └──────────────────┘
                    │
                    ▼
        ┌──────────────────────────────────────────┐
        │      Output Writers                      │
        │  • SQL Writer (INSERT statements)        │
        │  • COPY Writer (COPY format)             │
        │  • Streaming support                     │
        │  • Batch processing                      │
        └──────────────────────────────────────────┘
                              │
                              ▼
        ┌──────────────────────────────────────────┐
        │      Validation (Optional)               │
        │  • pg_query parser                       │
        │  • Syntax checking                       │
        │  • Error reporting                       │
        └──────────────────────────────────────────┘
```

### Key Components

#### **Schema Parser & Validator**
- Parses JSON schema definitions
- Validates types, constraints, and relationships
- Detects circular dependencies
- Provides detailed error messages with line numbers

#### **Generator Registry**
- Pluggable architecture for data generators
- Semantic detection based on column names (e.g., `email`, `phone_number`)
- Basic type generators (integer, varchar, timestamp, boolean)
- Custom pattern generators for business rules
- Time-series generators for temporal data

#### **Pipeline Coordinator**
- Resolves table dependencies (foreign keys)
- Determines optimal generation order
- Coordinates worker pools for parallel processing
- Tracks progress and statistics

#### **Worker Pool**
- Configurable parallel processing (1-100 workers)
- Context-aware cancellation
- Backpressure handling with buffered channels
- Error handling with continue-on-error option

#### **LRU Cache**
- Caches generated primary keys for foreign key lookups
- Thread-safe with read-write locks
- Configurable capacity (default: 10,000 entries)
- Tracks hits, misses, and evictions

#### **Output Writers**
- **SQL Writer**: Generates INSERT statements with batch support
- **COPY Writer**: Generates COPY format for faster loading
- Streaming architecture to handle large datasets
- Proper escaping and formatting

#### **Configuration System**
- Precedence: CLI flags > Environment variables > Config file > Defaults
- Viper-based configuration management
- YAML format (`.datagen.yaml`)

#### **Logging & Monitoring**
- Structured logging with zerolog
- Security event tracking
- Progress indicators with color support
- Verbose mode for debugging

---

## Project Structure

```
datagen-cli/
├── cmd/
│   └── datagen/              # CLI entry point
│       └── main.go
│
├── internal/                 # Internal packages
│   ├── cli/                  # CLI commands and infrastructure
│   │   ├── root.go          # Root command
│   │   ├── generate.go      # Generate command
│   │   ├── validate.go      # Validate command
│   │   ├── template.go      # Template command
│   │   ├── version.go       # Version command
│   │   ├── config.go        # Configuration management
│   │   ├── logging.go       # Structured logging
│   │   └── progress.go      # Progress indicators
│   │
│   ├── schema/              # Schema parsing and validation
│   │   ├── types.go         # Schema type definitions
│   │   ├── parser.go        # JSON schema parser
│   │   └── validator.go     # Schema validator
│   │
│   ├── generator/           # Data generation
│   │   ├── registry.go      # Generator registry
│   │   ├── context.go       # Generation context
│   │   ├── basic.go         # Basic type generators
│   │   ├── semantic.go      # Semantic generators
│   │   ├── sequence.go      # Sequence generators
│   │   └── cache.go         # LRU cache for FK lookups
│   │
│   ├── pipeline/            # Pipeline orchestration
│   │   ├── coordinator.go   # Pipeline coordinator
│   │   ├── dependency.go    # Dependency resolver
│   │   └── workers.go       # Worker pool
│   │
│   ├── pgdump/              # PostgreSQL dump format
│   │   ├── writer.go        # Base writer interface
│   │   ├── sql_writer.go    # SQL INSERT format
│   │   ├── copy_writer.go   # COPY format
│   │   ├── header.go        # Dump file headers
│   │   ├── helpers.go       # SQL helpers
│   │   └── validate.go      # SQL validation
│   │
│   └── templates/           # Pre-built templates
│       ├── registry.go      # Template registry
│       ├── ecommerce.go     # E-commerce template
│       ├── saas.go          # SaaS template
│       ├── healthcare.go    # Healthcare template
│       └── finance.go       # Finance template
│
├── tests/                   # Test suite
│   ├── unit/                # Unit tests
│   │   ├── cli/
│   │   ├── schema/
│   │   ├── generator/
│   │   ├── pipeline/
│   │   └── pgdump/
│   ├── integration/         # Integration tests
│   │   ├── pipeline/
│   │   └── templates/
│   └── benchmarks/          # Performance benchmarks
│
├── docs/                    # Documentation
│   ├── examples/            # Example schemas
│   └── man/                 # Man pages
│
├── scripts/                 # Build and utility scripts
│   ├── build.sh
│   └── test.sh
│
├── examples/                # Example schema files
│   ├── simple-schema.json
│   └── users-schema.json
│
├── specs/                   # Feature specifications
│   └── 001-json-to-pgdump/
│       ├── spec.md          # Feature specification
│       ├── plan.md          # Implementation plan
│       ├── tasks.md         # Task breakdown
│       ├── data-model.md    # Data model
│       ├── quickstart.md    # Quickstart guide
│       └── contracts/       # API contracts
│
├── .datagen.yaml.example    # Example configuration file
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── Makefile                 # Build automation
├── Dockerfile               # Container build
└── README.md                # This file
```

### Directory Purposes

- **`cmd/`**: Application entry points (CLI)
- **`internal/`**: Internal packages (not importable by external projects)
- **`tests/`**: All test files organized by type (unit, integration, benchmarks)
- **`docs/`**: User-facing documentation and examples
- **`scripts/`**: Build, test, and deployment scripts
- **`examples/`**: Sample schema files for users to try
- **`specs/`**: Detailed feature specifications and planning documents

---

## Configuration

datagen-cli can be configured via:

1. **Command-line flags** (highest priority)
2. **Environment variables** (prefix: `DATAGEN_`)
3. **Configuration file** (`.datagen.yaml`)
4. **Defaults** (lowest priority)

### Configuration File Example

Create `.datagen.yaml`:

```yaml
# General settings
verbose: false
default_seed: 0
default_format: sql
default_row_count: 1000
default_batch_size: 1000

# Performance
workers: 4
enable_cache: true
cache_size: 10000
stream_writes: true

# Logging
log_level: info        # debug, info, warn, error
log_format: text       # text, json
log_file: ""           # Empty = stderr
color_output: true
show_timestamp: true

# Progress
progress_bar: true
quiet_mode: false

# Output
json_output: false
pretty_print: true
```

### Environment Variables

All configuration options can be set via environment variables:

```bash
export DATAGEN_VERBOSE=true
export DATAGEN_WORKERS=8
export DATAGEN_LOG_LEVEL=debug
export DATAGEN_DEFAULT_FORMAT=copy
```

---

## Development

### Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile)
- Docker (optional, for integration tests)

### Build

```bash
# Build binary
make build

# Or with go directly
go build -o bin/datagen ./cmd/datagen
```

### Test

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific test package
go test ./internal/schema/...

# Run benchmarks
make benchmark
```

### Lint

```bash
# Run linters
make lint

# Auto-fix issues
make lint-fix
```

### Code Quality

The project enforces:
- ≥90% test coverage
- Maximum function length: 20 lines
- Cyclomatic complexity ≤ 10
- No duplicate code
- Security scanning with govulncheck

### Project Stats

- **Test Coverage**: 97.5%
- **Total Tasks**: 120
- **Completed Tasks**: 117 (97.5%)
- **Lines of Code**: ~10,000
- **Test Files**: 50+

---

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`make test`)
5. Run linters (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Style

- Follow Go best practices and idioms
- Write clear, descriptive commit messages
- Add tests for all new functionality
- Update documentation as needed
- Keep functions small and focused (≤20 lines)

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [gofakeit](https://github.com/brianvoe/gofakeit) - Fake data generation
- [pg_query_go](https://github.com/pganalyze/pg_query_go) - PostgreSQL query parser
- [zerolog](https://github.com/rs/zerolog) - Structured logging
- [testify](https://github.com/stretchr/testify) - Testing toolkit

---

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/NhaLeTruc/datagen-cli/issues)
- 💬 [Discussions](https://github.com/NhaLeTruc/datagen-cli/discussions)

---

**Made with ❤️ by the datagen-cli team**
