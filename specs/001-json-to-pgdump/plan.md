# Implementation Plan: JSON Schema to PostgreSQL Dump Generator

**Branch**: `001-json-to-pgdump` | **Date**: 2025-11-15 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-json-to-pgdump/spec.md`

## Summary

Build a CLI tool in Go that transforms declarative JSON schema definitions into fully-formed PostgreSQL dump files containing realistic mock data, without requiring a running PostgreSQL instance. The tool uses a pipeline architecture with streaming support for large datasets, intelligent context-aware data generation, and supports multiple output formats. Core value proposition: enable developers to create realistic test databases quickly while maintaining 100% PostgreSQL compatibility and avoiding the privacy/compliance risks of using production data.

**Technical Approach**: Pipeline architecture with Go 1.21+, using Cobra for CLI, gofakeit for data generation, custom SQL builder for PostgreSQL dump format, and concurrent worker pools for performance. Emphasis on TDD with ≥90% coverage, security-first input validation, and streaming architecture to handle datasets up to 100GB.

## Technical Context

**Language/Version**: Go 1.21+ (supports generics, improved performance, security patches)
**Primary Dependencies**:
- CLI: `spf13/cobra` v1.8+ (command structure), `spf13/viper` v1.18+ (config management)
- Data Generation: `brianvoe/gofakeit/v6` v6.28+ (fake data), custom generator registry
- PostgreSQL: `pganalyze/pg_query_go/v5` (SQL validation), custom dump format writer
- Testing: `stretchr/testify` v1.9+ (assertions/mocks), `testcontainers-go` (PostgreSQL integration tests)

**Storage**: File-based (no database required) - reads JSON schemas, writes PostgreSQL dump files (.dump, .sql, .copy formats)

**Testing**: Go native `testing` package with testify assertions, table-driven tests for generators, Docker-based PostgreSQL validation tests, property-based testing for data distributions, golden file regression tests

**Target Platform**: Cross-platform CLI (Linux amd64/arm64, macOS amd64/arm64, Windows amd64) - distributed via GitHub releases, Homebrew, APT/YUM, Chocolatey, Docker

**Project Type**: Single Go binary CLI tool

**Performance Goals**:
- Generate 1GB dump in <30 seconds (33MB/s throughput)
- Handle 1M rows across multiple tables in <2 minutes
- Memory usage <500MB for datasets up to 100K rows
- Startup time <100ms for basic commands

**Constraints**:
- No runtime PostgreSQL instance required
- Streaming architecture (no full dataset in memory)
- Deterministic output with seeds (byte-identical across runs)
- 100% pg_restore compatibility (PostgreSQL 12-16)
- ≥90% test coverage (constitution requirement)

**Scale/Scope**:
- Support schemas with 100+ tables
- Generate datasets up to 100GB
- Handle 10M+ rows with streaming
- Support 200+ semantic column patterns (email, phone, address, etc.)
- 4 pre-built templates (ecommerce, SaaS, healthcare, finance)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development (NON-NEGOTIABLE)
- ✅ **PASS**: Plan includes comprehensive testing strategy (unit, integration, validation, performance, fuzzing)
- ✅ **PASS**: TDD workflow enforced via constitution - tests written before implementation
- ✅ **PASS**: ≥90% coverage target specified in constraints
- ✅ **PASS**: Integration tests validate PostgreSQL archive compatibility (constitution requirement)
- **Action**: Every task in tasks.md must include test creation BEFORE implementation

### II. Clean Code Standards
- ✅ **PASS**: Go's strong typing enforces type safety
- ✅ **PASS**: Architecture uses SOLID principles (single responsibility per pipeline stage, interface-based generator registry)
- ⚠️ **WATCH**: Must enforce max 20 lines/function, cyclomatic complexity ≤10 via linting
- **Action**: Configure golangci-lint with funlen, gocyclo, dupl linters

### III. Security-First Architecture
- ✅ **PASS**: Input validation via JSON schema validation (go-jsonschema)
- ✅ **PASS**: Path sanitization for file operations (filepath.Clean, restrict to working dir)
- ✅ **PASS**: SQL injection prevention via template-based generation (no string concatenation)
- ✅ **PASS**: Dependency scanning planned (GitHub Actions, govulncheck)
- ⚠️ **WATCH**: Audit logging for file operations needs implementation details
- **Action**: Add structured logging (zerolog) with security event categorization

### IV. PostgreSQL Archive Integrity
- ✅ **PASS**: Custom dump format writer ensuring pg_dump compatibility
- ✅ **PASS**: Testing against PostgreSQL 12-16 specified
- ✅ **PASS**: Validation via pg_restore in integration tests
- ✅ **PASS**: Proper metadata, headers, compression support planned
- **Action**: Phase 1 must define dump file format specification in contracts/

### V. CLI Interface Contract
- ✅ **PASS**: Cobra framework provides POSIX-compliant CLI
- ✅ **PASS**: Stdin/stdout/stderr protocol specified
- ✅ **PASS**: Exit codes, --help, --version, progress indicators planned
- ✅ **PASS**: JSON and human-readable output formats supported
- **Action**: Define exact CLI contract in contracts/cli.md

### VI. Performance & Scale
- ✅ **PASS**: Streaming architecture prevents memory exhaustion
- ✅ **PASS**: Performance targets defined (1GB in <30s, 1M rows in <2min)
- ✅ **PASS**: Worker pool pattern for concurrent generation
- ✅ **PASS**: Memory constraints specified (<500MB for 100K rows)
- **Action**: Add performance benchmarks to test suite

**Constitution Gate Result**: ✅ **PASSED** - All critical requirements met, minor actions identified for implementation phase

## Project Structure

### Documentation (this feature)

```text
specs/001-json-to-pgdump/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── cli.md          # CLI command interface specification
│   ├── schema-format.md # JSON schema format specification
│   └── dump-format.md  # PostgreSQL dump format specification
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
└── datagen/
    └── main.go          # CLI entry point

internal/
├── cli/                 # CLI commands and flags
│   ├── root.go         # Root command
│   ├── generate.go     # Generate command
│   ├── validate.go     # Validate command
│   └── version.go      # Version command
├── schema/             # Schema parsing and validation
│   ├── parser.go       # JSON schema parser
│   ├── validator.go    # Schema validator
│   └── types.go        # Schema type definitions
├── generator/          # Data generation
│   ├── registry.go     # Generator registry
│   ├── semantic.go     # Semantic column detection
│   ├── basic.go        # Basic type generators
│   ├── custom.go       # Custom pattern generators
│   └── timeseries.go   # Time-series generators
├── pgdump/             # PostgreSQL dump format
│   ├── writer.go       # Dump file writer
│   ├── header.go       # Dump file headers
│   ├── toc.go          # Table of contents
│   └── compression.go  # Compression support
├── pipeline/           # Pipeline orchestration
│   ├── coordinator.go  # Pipeline coordinator
│   ├── workers.go      # Worker pool
│   └── dependency.go   # Dependency resolver
└── templates/          # Built-in templates
    ├── ecommerce.json
    ├── saas.json
    ├── healthcare.json
    └── finance.json

pkg/                     # Public API (if needed for plugins)
└── generator/
    └── interface.go     # Public generator interface

tests/
├── unit/               # Unit tests (70%)
│   ├── schema/
│   ├── generator/
│   └── pgdump/
├── integration/        # Integration tests (25%)
│   ├── pipeline/
│   └── postgresql/     # Docker-based PG tests
└── e2e/                # End-to-end tests (5%)
    └── scenarios/

scripts/
├── build.sh            # Build script
├── test.sh             # Test runner
└── release.sh          # Release automation

docs/
├── architecture.md     # Architecture documentation
├── generators.md       # Generator documentation
└── examples/           # Example schemas
```

**Structure Decision**: Single Go project structure following standard Go layout (cmd/, internal/, pkg/, tests/). Using internal/ for non-exported packages ensures clean boundaries. The pipeline architecture maps cleanly to package structure: schema parsing → generation → dump writing. Templates embedded in binary using go:embed for portability.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations requiring justification. All constitutional principles are satisfied by the planned architecture.

## Phase 0: Research & Technology Decisions

### Research Tasks

The user has provided detailed architectural guidance. Key research areas to validate:

1. **PostgreSQL Dump Format Specification**
   - Research pg_dump custom format internals
   - Understand TOC (Table of Contents) structure
   - Document compression algorithms (gzip, custom)
   - Identify version-specific format differences (PG 12-16)

2. **Go Library Evaluation**
   - Validate gofakeit v6 capabilities for semantic data generation
   - Evaluate pg_query_go for SQL validation without database
   - Research LRU cache implementations for foreign key lookups
   - Investigate go:embed for template bundling

3. **Concurrency Patterns**
   - Worker pool best practices for dependency-aware generation
   - Channel patterns for backpressure and memory management
   - sync.Pool usage for object reuse in high-throughput scenarios

4. **Testing Strategy**
   - testcontainers-go setup for PostgreSQL integration tests
   - Property-based testing libraries for data distribution validation
   - Golden file testing for SQL output regression tests

### Output Artifact

`research.md` will contain:
- Decision rationale for each technology choice
- Alternatives considered and rejected
- Best practices discovered
- Risk mitigation strategies
- Performance benchmarking results (if applicable)

## Phase 1: Design Artifacts

### data-model.md

Will define internal domain models:
- **Schema**: Top-level schema definition with tables, types, sequences
- **Table**: Table specification with columns, constraints, dependencies
- **Column**: Column definition with type, constraints, generation rules
- **Generator**: Data generator interface and implementations
- **DumpFile**: Output file structure with headers, TOC, data sections
- **Template**: Pre-built schema template structure

Each entity includes:
- Fields with types
- Validation rules from spec requirements
- Relationships to other entities
- State transitions (e.g., Schema: parsing → validated → generated)

### contracts/

Will contain:

**cli.md** - CLI interface specification:
```bash
datagen generate <schema.json> [flags]
datagen validate <schema.json>
datagen template <template-name> [flags]
datagen version

Flags:
  --output, -o <file>       Output file path (default: stdout)
  --format <format>         Output format: custom|sql|copy (default: custom)
  --seed <value>            Random seed for deterministic output
  --rows <count>            Default row count per table
  --config <file>           Configuration file path
  --verbose, -v             Verbose output
  --dry-run                 Validate only, don't generate
```

**schema-format.md** - JSON schema format with examples

**dump-format.md** - PostgreSQL dump file format specification

### quickstart.md

Will provide:
1. Installation instructions (go install, homebrew, binary download)
2. Basic usage example (simple 2-table schema)
3. Advanced example (using templates, custom generators)
4. Testing the output (pg_restore walkthrough)
5. Common troubleshooting

## Phase 2: Task Generation

Not executed by this command - use `/speckit.tasks` after plan completion.

## Next Steps

1. ✅ Constitution check passed - proceed to Phase 0
2. Generate research.md with technology validation
3. Generate Phase 1 artifacts (data-model.md, contracts/, quickstart.md)
4. Update agent context files
5. Ready for `/speckit.tasks` to generate implementation tasks