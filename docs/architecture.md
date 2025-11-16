# Architecture Documentation

**Project**: datagen-cli - JSON Schema to PostgreSQL Dump Generator
**Version**: 1.0
**Language**: Go 1.21+
**Last Updated**: 2025-11-16

## Table of Contents

- [System Overview](#system-overview)
- [Architecture Patterns](#architecture-patterns)
- [Component Architecture](#component-architecture)
- [Data Flow](#data-flow)
- [Design Decisions](#design-decisions)
- [Extension Points](#extension-points)
- [Performance Considerations](#performance-considerations)
- [Security Architecture](#security-architecture)

## System Overview

### Purpose

datagen-cli is a command-line tool that generates PostgreSQL dump files containing realistic mock data from declarative JSON schema definitions, without requiring a running PostgreSQL instance.

### Core Value Proposition

- **No Database Required**: Generate PostgreSQL dumps without running PostgreSQL
- **Realistic Data**: Context-aware generation based on semantic column names
- **100% Compatible**: Output works with pg_restore for PostgreSQL 12-16
- **Deterministic**: Same seed produces byte-identical output
- **High Performance**: Streaming architecture handles datasets up to 100GB

### Architecture Style

**Pipeline Architecture** with:
- **Streaming data flow** to minimize memory footprint
- **Registry pattern** for pluggable data generators
- **Worker pools** for concurrent data generation
- **Dependency resolution** for foreign key relationships

## Architecture Patterns

### 1. Pipeline Pattern

The core workflow follows a pipeline architecture:

```
JSON Schema → Parse → Validate → Resolve Dependencies → Generate Data → Write Dump
```

Each stage is independent and testable:

| Stage | Input | Output | Responsibility |
|-------|-------|--------|----------------|
| Parse | JSON bytes | Schema struct | Parse JSON, validate syntax |
| Validate | Schema struct | Errors or nil | Validate types, constraints, references |
| Resolve | Schema struct | Sorted table list | Topological sort based on FKs |
| Generate | Table spec | Row data | Generate realistic data per column |
| Write | Row data | Dump file | Format as PostgreSQL dump |

**Implementation**: `internal/pipeline/coordinator.go`

### 2. Registry Pattern

Data generators use a registry pattern for extensibility:

```go
type Generator interface {
    Generate(ctx *GenerationContext, config map[string]interface{}) (interface{}, error)
    Type() string
}

// Global registry
var registry = make(map[string]Generator)

func Register(gen Generator) {
    registry[gen.Type()] = gen
}
```

**Benefits**:
- Easy to add new generator types
- Generators can be registered at package init time
- Enables user-defined custom generators (future extension)

**Implementation**: `internal/generator/registry.go`

### 3. Strategy Pattern (Generator Selection)

Generator selection follows a priority-based strategy:

1. **Explicit generator** specified in schema (`generator: "weighted_enum"`)
2. **Semantic detection** based on column name (`email` → EmailGenerator)
3. **Type-based default** based on PostgreSQL type (`varchar` → VarcharGenerator)

**Implementation**: `internal/pipeline/coordinator.go`, `internal/generator/semantic.go`

### 4. Streaming Architecture

To handle large datasets without loading everything into memory:

- **Batch processing**: Write rows in batches of 1000
- **Channel-based flow**: Use Go channels for backpressure
- **No full dataset buffering**: Stream directly to output file

**Implementation**: `internal/pgdump/writer.go`

## Component Architecture

### CLI Layer (`internal/cli/`)

**Responsibility**: Command-line interface and user interaction

**Components**:
- `root.go`: Root command setup, global flags, Cobra initialization
- `generate.go`: Generate command implementation (main entry point)
- `validate.go`: Schema validation command
- `template.go`: Template management commands
- `version.go`: Version information command

**Dependencies**: Cobra (CLI framework), Viper (config management)

**Key Design**:
- Each command is a separate file implementing `cobra.Command`
- Commands orchestrate pipeline but contain no business logic
- Error handling follows CLI conventions (exit codes, stderr)

### Schema Layer (`internal/schema/`)

**Responsibility**: JSON schema parsing and validation

**Components**:
- `types.go`: Schema, Table, Column, ForeignKey structs
- `parser.go`: JSON unmarshaling and schema object construction
- `validator.go`: Schema validation rules and error messages

**Validation Rules**:
- PostgreSQL type validation (validates against known PG types)
- Foreign key reference validation (table and column existence)
- Circular dependency detection (DFS cycle detection)
- Constraint validation (unique, check, indexes)

**Key Design**:
- Immutable schema objects after parsing
- Validation returns all errors, not just first failure
- Error messages include suggestions for common mistakes

### Generator Layer (`internal/generator/`)

**Responsibility**: Realistic data generation for all column types

**Components**:
- `registry.go`: Generator registration and lookup
- `context.go`: Generation context with seeded random
- `basic.go`: Basic PostgreSQL type generators (integer, varchar, timestamp, boolean)
- `semantic.go`: Semantic generators (email, phone, name, address)
- `custom.go`: Custom pattern generators (weighted enum, regex pattern, templates)
- `timeseries.go`: Time-series generators with patterns (uniform, business hours, daily peaks)
- `sequence.go`: Sequence generator for serial/bigserial types

**Generator Types**:

| Category | Generators | Use Case |
|----------|------------|----------|
| Basic | Integer, Varchar, Text, Timestamp, Boolean, UUID, JSON | Default type-based generation |
| Semantic | Email, Phone, Name, Address, City, Country, PostalCode | Intelligent column name detection |
| Custom | WeightedEnum, Pattern, Template, IntegerRange | User-specified business rules |
| Timeseries | Uniform, BusinessHours, DailyPeak | Time-series data with patterns |
| Special | Serial, Bigserial, ForeignKey | PostgreSQL-specific types |

**Key Design**:
- All generators use GenerationContext with seeded rand (deterministic)
- Generators are stateless (no internal state)
- Thread-safe for concurrent use in worker pools

### PostgreSQL Dump Layer (`internal/pgdump/`)

**Responsibility**: Write PostgreSQL-compatible dump files

**Components**:
- `writer.go`: Main dump writer with format selection (factory pattern)
- `sql_writer.go`: SQL INSERT format writer
- `copy_writer.go`: COPY format writer
- `sql_escape.go`: SQL string escaping and quoting
- `copy_escape.go`: COPY format data escaping (TSV)

**Output Formats**:

| Format | File Extension | Use Case | Implementation |
|--------|---------------|----------|----------------|
| SQL | `.sql` | Human-readable, debugging | INSERT statements with batching |
| COPY | `.copy.sql` | Fast imports, large datasets | COPY FROM stdin with TSV data |

**Key Design**:
- Factory pattern for format selection
- Streaming write (no full dump in memory)
- Proper escaping for all PostgreSQL data types
- CREATE TABLE statements include all constraints

### Pipeline Layer (`internal/pipeline/`)

**Responsibility**: Orchestrate generation workflow and dependencies

**Components**:
- `coordinator.go`: Main pipeline orchestrator
- `dependency.go`: Dependency resolution and topological sort

**Pipeline Flow**:

1. **Parse Schema**: JSON → Schema struct
2. **Validate Schema**: Check types, constraints, references
3. **Resolve Dependencies**: Topological sort based on foreign keys
4. **For Each Table** (in dependency order):
   - Select generator for each column
   - Generate rows in batches
   - Write to output file
5. **Finalize**: Close file, verify integrity

**Key Design**:
- Tables generated in dependency order (dependencies first)
- Self-referencing foreign keys allowed (handled specially)
- Circular dependencies detected and rejected during validation

### Templates Layer (`internal/templates/`)

**Responsibility**: Pre-built schema templates for common use cases

**Components**:
- `embed.go`: Embedded template files using go:embed
- `ecommerce.json`: E-commerce schema (products, orders, customers)
- `saas.json`: SaaS schema (tenants, users, subscriptions)
- `healthcare.json`: Healthcare schema (patients, appointments, records)
- `finance.json`: Finance schema (accounts, transactions, portfolios)

**Key Design**:
- Templates embedded in binary (no external files needed)
- Templates parameterizable (override row counts, customize values)
- Templates demonstrate best practices for schema design

## Data Flow

### Generate Command Flow

```
User Input
    ↓
┌───────────────────────────────────────┐
│ CLI Layer (generate.go)               │
│ - Parse flags                         │
│ - Read input (stdin or file)          │
│ - Select output format                │
└───────────────────────────────────────┘
    ↓
┌───────────────────────────────────────┐
│ Schema Layer (parser.go)              │
│ - Parse JSON                          │
│ - Build Schema struct                 │
└───────────────────────────────────────┘
    ↓
┌───────────────────────────────────────┐
│ Schema Layer (validator.go)           │
│ - Validate types                      │
│ - Check constraints                   │
│ - Detect circular dependencies        │
└───────────────────────────────────────┘
    ↓
┌───────────────────────────────────────┐
│ Pipeline Layer (dependency.go)        │
│ - Build dependency graph              │
│ - Topological sort                    │
│ - Return ordered table list           │
└───────────────────────────────────────┘
    ↓
┌───────────────────────────────────────┐
│ Pipeline Layer (coordinator.go)       │
│ For each table:                       │
│   - Select generators per column      │
│   - Generate batch of rows            │
│   - Stream to writer                  │
└───────────────────────────────────────┘
    ↓
┌───────────────────────────────────────┐
│ Generator Layer (*.go)                │
│ - Use seeded random                   │
│ - Generate realistic data             │
│ - Respect constraints                 │
└───────────────────────────────────────┘
    ↓
┌───────────────────────────────────────┐
│ PGDump Layer (writer.go)              │
│ - Format as SQL/COPY                  │
│ - Escape data properly                │
│ - Write to file/stdout                │
└───────────────────────────────────────┘
    ↓
PostgreSQL Dump File
```

### Dependency Resolution Example

Given schema:
```
users (no FKs)
posts (FK → users)
comments (FK → posts, FK → users)
```

Dependency graph:
```
users
  ↓
posts ← users
  ↓
comments ← posts, users
```

Generation order: `[users, posts, comments]`

**Algorithm**: Kahn's algorithm for topological sort (DFS with cycle detection)

## Design Decisions

### 1. Why Go?

**Decision**: Use Go 1.21+ for implementation

**Rationale**:
- Strong typing and compile-time safety
- Excellent concurrency primitives (goroutines, channels)
- Single binary distribution (no runtime dependencies)
- Cross-platform compilation built-in
- Fast compilation and execution
- Rich standard library (especially for I/O and parsing)

**Alternatives Considered**:
- Python: Slower, requires runtime, harder to distribute
- Rust: Steeper learning curve, longer compile times
- Node.js: Requires runtime, harder to control memory

### 2. Why Custom Format Writers Instead of pg_dump Library?

**Decision**: Implement custom SQL and COPY format writers

**Rationale**:
- No Go library exists for writing pg_dump custom format
- Custom format is complex and version-specific
- SQL and COPY formats are simpler and well-documented
- SQL format is human-readable (better for debugging)
- COPY format is fast and widely supported

**Trade-off**: Custom format support deferred to future version

### 3. Why Registry Pattern for Generators?

**Decision**: Use registry pattern with init-time registration

**Rationale**:
- Enables modular generator development
- Easy to add new generators without modifying core
- Supports future plugin system
- Clean separation of concerns

**Implementation**:
```go
func init() {
    generator.Register(&EmailGenerator{})
    generator.Register(&PhoneGenerator{})
}
```

### 4. Why Streaming Architecture?

**Decision**: Stream data directly to output, avoid buffering full dataset

**Rationale**:
- Enables datasets larger than available memory
- Constant memory usage regardless of dataset size
- Better performance (no unnecessary buffering)

**Trade-off**: Harder to implement multi-pass operations

### 5. Why Semantic Detection?

**Decision**: Automatically detect column semantics from names (email, phone, etc.)

**Rationale**:
- Reduces schema verbosity
- Makes generated data realistic by default
- Common convention across industries

**Example**:
```json
{"name": "email", "type": "varchar(255)"}
// Automatically uses EmailGenerator
```

### 6. Why Embed Templates?

**Decision**: Embed templates in binary using go:embed

**Rationale**:
- Single binary distribution (no external files)
- Templates always available
- Simpler deployment

**Implementation**:
```go
//go:embed *.json
var templates embed.FS
```

## Extension Points

### 1. Custom Generators

**How to Add**: Implement Generator interface and register

```go
type CustomGenerator struct{}

func (g *CustomGenerator) Generate(ctx *GenerationContext, config map[string]interface{}) (interface{}, error) {
    // Custom logic
}

func (g *CustomGenerator) Type() string {
    return "custom_type"
}

func init() {
    generator.Register(&CustomGenerator{})
}
```

### 2. New Output Formats

**How to Add**: Implement Writer interface

```go
type Writer interface {
    WriteHeader(schema *Schema) error
    WriteTable(table *Table, rows []Row) error
    Close() error
}
```

### 3. New Templates

**How to Add**: Add JSON file to `internal/templates/` and rebuild

### 4. Custom Validation Rules

**How to Add**: Extend `schema.Validate()` function

```go
func Validate(s *Schema) []error {
    var errs []error
    // Add custom validation logic
    return errs
}
```

## Performance Considerations

### Memory Management

**Target**: <500MB for datasets up to 100K rows

**Strategies**:
- Batch processing (1000 rows at a time)
- Streaming I/O (no full file buffering)
- Reuse buffers (sync.Pool for string builders)
- Limit FK lookup cache size (LRU with max 10K entries)

### Concurrency

**Not Yet Implemented** (T102-T107):
- Worker pool for parallel table generation
- Channel-based backpressure
- Configurable worker count (--jobs flag)

**Future Optimization**:
```go
// Worker pool pattern
type WorkerPool struct {
    jobs    chan *TableJob
    results chan *TableResult
    workers int
}
```

### Streaming Write

**Current**: Batch writes every 1000 rows
**Future**: Configurable batch size based on row complexity

## Security Architecture

### Input Validation

**Schema Validation**:
- Validate all PostgreSQL types against whitelist
- Reject unknown types with suggestions
- Check all foreign key references
- Detect circular dependencies

**Path Sanitization**:
```go
// Ensure paths are within working directory
cleanPath := filepath.Clean(userPath)
if !strings.HasPrefix(cleanPath, workingDir) {
    return errors.New("path outside working directory")
}
```

### SQL Injection Prevention

**Never use string concatenation** for SQL generation:

```go
// GOOD: Template-based generation
fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (%s);\n",
    quoteIdentifier(table),
    quoteIdentifierList(columns),
    quoteLiteralList(values))

// BAD: String concatenation (never used)
sql := "INSERT INTO " + table + " VALUES (" + values + ")"
```

### Escaping

**SQL Format**:
- Identifiers: Double-quote and escape internal quotes
- Strings: Single-quote and escape with `E'...'` for special chars
- NULL: Literal `NULL` keyword

**COPY Format**:
- Tab delimiter: Escape as `\t`
- Newline: Escape as `\n`
- Backslash: Escape as `\\`
- NULL: Literal `\N`

### Dependency Scanning

**CI/CD**: GitHub Actions runs `govulncheck` on every commit

## Testing Architecture

### Test Pyramid

- **Unit Tests (70%)**: `tests/unit/` - Fast, isolated, no I/O
- **Integration Tests (25%)**: `tests/integration/` - Real file I/O, in-memory validation
- **End-to-End Tests (5%)**: `tests/e2e/` - Full CLI workflows

### Test-Driven Development (TDD)

**Process**:
1. Write test first (Red phase)
2. Verify test fails
3. Implement feature (Green phase)
4. Verify test passes
5. Refactor (Refactor phase)

**Example**:
```go
// 1. Write test first
func TestEmailGenerator(t *testing.T) {
    gen := &EmailGenerator{}
    ctx := NewGenerationContext(12345)

    email, err := gen.Generate(ctx, nil)
    require.NoError(t, err)
    assert.Regexp(t, `^[^@]+@[^@]+\.[^@]+$`, email)
}

// 2. Implement generator
type EmailGenerator struct{}
func (g *EmailGenerator) Generate(ctx *GenerationContext, config map[string]interface{}) (interface{}, error) {
    return gofakeit.Email(), nil
}
```

### Coverage Requirements

**Target**: ≥90% code coverage
**Enforcement**: CI fails if coverage drops below 90%

**Coverage by Package**:
- `internal/schema`: 100% (critical path)
- `internal/generator`: ≥95% (many edge cases)
- `internal/pgdump`: ≥90% (format variations)
- `internal/pipeline`: ≥90% (orchestration logic)
- `internal/cli`: ≥80% (harder to test, UI logic)

## Appendix: Key Files Reference

| File | Purpose | Key Functions/Types |
|------|---------|---------------------|
| `internal/cli/generate.go` | Main CLI command | `NewGenerateCommand()`, flag parsing |
| `internal/schema/parser.go` | JSON parsing | `Parse()` |
| `internal/schema/validator.go` | Validation | `Validate()`, `validateForeignKey()` |
| `internal/schema/types.go` | Data structures | `Schema`, `Table`, `Column` |
| `internal/generator/registry.go` | Generator registry | `Register()`, `Get()` |
| `internal/generator/semantic.go` | Semantic detection | `DetectSemantic()`, email/phone generators |
| `internal/generator/custom.go` | Custom generators | `WeightedEnumGenerator`, `PatternGenerator` |
| `internal/pipeline/coordinator.go` | Pipeline orchestration | `Generate()`, table generation loop |
| `internal/pipeline/dependency.go` | Dependency resolution | `ResolveDependencies()`, topological sort |
| `internal/pgdump/sql_writer.go` | SQL format writer | `WriteSQLDump()`, INSERT statement generation |
| `internal/pgdump/copy_writer.go` | COPY format writer | `WriteCOPYDump()`, TSV data formatting |

## Version History

- **v1.0** (2025-11-16): Initial architecture documentation
