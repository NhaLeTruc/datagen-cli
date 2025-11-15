---

description: "Task list for JSON Schema to PostgreSQL Dump Generator"
---

# Tasks: JSON Schema to PostgreSQL Dump Generator

**Input**: Design documents from `/specs/001-json-to-pgdump/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), data-model.md, contracts/

**Tests**: TDD is MANDATORY per constitution - tests MUST be written BEFORE implementation code

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single Go CLI project**: `cmd/`, `internal/`, `tests/` at repository root
- All paths are from repository root

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize Go module with `go mod init github.com/NhaLeTruc/datagen-cli`
- [x] T002 Create directory structure: cmd/datagen/, internal/{cli,schema,generator,pgdump,pipeline,templates}/, tests/{unit,integration,e2e}/
- [x] T003 [P] Create go.mod with dependencies: cobra v1.8+, viper v1.18+, gofakeit/v6 v6.28+, testify v1.9+
- [x] T004 [P] Configure golangci-lint with funlen, gocyclo, dupl linters in .golangci.yml
- [x] T005 [P] Create Makefile with targets: build, test, lint, coverage, clean
- [x] T006 [P] Setup GitHub Actions workflow in .github/workflows/ci.yml for: lint, test, coverage
- [x] T007 [P] Create .gitignore for Go projects (bin/, vendor/, *.dump, *.sql)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Foundational: Schema Foundation

- [x] T008 Write unit test for schema types in tests/unit/schema/types_test.go (TDD: test Schema, Table, Column structs)
- [x] T009 Implement schema type definitions in internal/schema/types.go (Schema, Table, Column, ForeignKey, etc.)
- [x] T010 Write unit test for JSON schema parser in tests/unit/schema/parser_test.go (TDD: valid/invalid JSON, edge cases)
- [x] T011 Implement JSON schema parser in internal/schema/parser.go (parse JSON to Schema struct)
- [x] T012 Write unit test for schema validator in tests/unit/schema/validator_test.go (TDD: type validation, FK references, circular deps)
- [x] T013 Implement schema validator in internal/schema/validator.go (validate types, constraints, references)

### Foundational: CLI Framework

- [x] T014 [P] Write unit test for root command in tests/unit/cli/root_test.go (TDD: help text, flags parsing)
- [x] T015 [P] Implement root command in internal/cli/root.go with Cobra (--help, --version, --config)
- [x] T016 [P] Write unit test for version command in tests/unit/cli/version_test.go (TDD: version output formats)
- [x] T017 [P] Implement version command in internal/cli/version.go (show version, build info, Go version)

### Foundational: Generator Registry

- [x] T018 Write unit test for generator interface in tests/unit/generator/registry_test.go (TDD: register, get, not found cases)
- [x] T019 Implement generator registry in internal/generator/registry.go (registry pattern, thread-safe)
- [x] T020 Write unit test for generation context in tests/unit/generator/context_test.go (TDD: context creation, seeding)
- [x] T021 Implement generation context in internal/generator/context.go (GenerationContext with seeded rand)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Basic Schema to Dump Conversion (Priority: P1) 🎯 MVP

**Goal**: Enable developers to transform a simple JSON schema (2-3 tables) into a valid PostgreSQL dump file that can be restored with pg_restore

**Independent Test**: Create JSON schema with users and posts tables, generate dump, restore to PostgreSQL, verify tables and data exist

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T022 [P] [US1] Contract test for basic type generators in tests/unit/generator/basic_test.go (TDD: integer, varchar, timestamp generators)
- [x] T023 [P] [US1] Integration test for simple schema pipeline in tests/integration/pipeline/basic_pipeline_test.go (TDD: 2-table schema, verify row counts)
- [ ] T024 [P] [US1] Integration test for PostgreSQL restore in tests/integration/postgresql/restore_test.go (TDD: generate dump, restore, query data)

### Implementation for User Story 1

#### Basic Generators

- [x] T025 [P] [US1] Implement basic type generators in internal/generator/basic.go (integer, varchar, text, timestamp, boolean)
- [x] T026 [P] [US1] Implement sequence generator in internal/generator/sequence.go (for serial/bigserial columns)
- [x] T027 [US1] Register basic generators in internal/generator/registry.go (register all basic types on init)

#### PostgreSQL Dump Writer (Custom Format)

- [x] T028 [P] [US1] Write unit test for dump header in tests/unit/pgdump/header_test.go (TDD: magic bytes, version, metadata)
- [x] T029 [P] [US1] Implement dump header writer in internal/pgdump/header.go (write PGDMP header, version, timestamp)
- [x] T03- [x] T030 [P] [US1] Write unit test for TOC entry in tests/unit/pgdump/toc_test.go (TDD: entry creation, dependencies, offsets)
- [x] T03- [x] T031 [P] [US1] Implement TOC structure in internal/pgdump/toc.go (TOCEntry, dependency tracking)
- [x] T03- [x] T032 [US1] Write unit test for dump writer in tests/unit/pgdump/writer_test.go (TDD: full dump flow, compression)
- [x] T03- [x] T033 [US1] Implement dump file writer in internal/pgdump/writer.go (streaming write, TOC, data sections)
- [x] T03- [x] T034 [P] [US1] Write unit test for compression in tests/unit/pgdump/compression_test.go (TDD: gzip compression, level 6)
- [x] T03- [x] T035 [P] [US1] Implement gzip compression in internal/pgdump/compression.go (compress data sections)

#### Pipeline Orchestration

- [x] T03- [x] T036 [US1] Write unit test for dependency resolver in tests/unit/pipeline/dependency_test.go (TDD: topological sort, detect cycles)
- [x] T03- [x] T037 [US1] Implement dependency resolver in internal/pipeline/dependency.go (resolve FK dependencies, sort tables)
- [x] T03- [x] T038 [US1] Write unit test for pipeline coordinator in tests/unit/pipeline/coordinator_test.go (TDD: schema → dump flow)
- [x] T03- [x] T039 [US1] Implement pipeline coordinator in internal/pipeline/coordinator.go (orchestrate: parse → validate → generate → write)

#### Generate Command

- [x] T04- [x] T040 [US1] Write unit test for generate command in tests/unit/cli/generate_test.go (TDD: flags, stdin/stdout, file paths)
- [x] T04- [x] T041 [US1] Implement generate command in internal/cli/generate.go (parse args, call pipeline, handle errors)
- [x] T04- [x] T042 [US1] Wire generate command to root in internal/cli/root.go

#### CLI Entry Point

- [x] T04- [x] T043 [US1] Implement main.go in cmd/datagen/main.go (initialize CLI, execute root command)

**Checkpoint**: At this point, User Story 1 should be fully functional - can generate and restore basic dumps

---

## Phase 4: User Story 2 - Intelligent Context-Aware Data Generation (Priority: P2)

**Goal**: Automatically generate realistic, properly formatted data for columns with semantic names (email, phone, first_name, postal_code)

**Independent Test**: Define schema with semantic column names, generate dump, verify emails are valid, phone numbers formatted, names realistic

### Tests for User Story 2 ⚠️

- [x] T044 [P] [US2] Unit test for semantic detection in tests/unit/generator/semantic_test.go (TDD: detect email, phone, name, address patterns)
- [x] T045 [P] [US2] Contract test for semantic generators in tests/unit/generator/semantic_generators_test.go (TDD: email, phone, name, address outputs)
- [x] T046 [P] [US2] Integration test for semantic data generation in tests/integration/pipeline/semantic_pipeline_test.go (TDD: verify realistic data patterns)

### Implementation for User Story 2

- [x] T047 [P] [US2] Implement semantic column detector in internal/generator/semantic.go (pattern matching for column names)
- [x] T048 [P] [US2] Implement email generator using gofakeit in internal/generator/semantic.go (EmailGenerator)
- [x] T049 [P] [US2] Implement phone generator using gofakeit in internal/generator/semantic.go (PhoneGenerator)
- [x] T050 [P] [US2] Implement name generators using gofakeit in internal/generator/semantic.go (FirstName, LastName, FullName)
- [x] T051 [P] [US2] Implement address generators using gofakeit in internal/generator/semantic.go (Address, City, State, Country, PostalCode)
- [x] T052 [P] [US2] Implement timestamp generators in internal/generator/semantic.go (CreatedAt, UpdatedAt with realistic ranges)
- [x] T053 [US2] Register semantic generators in internal/generator/registry.go (register all semantic generators)
- [x] T054 [US2] Update pipeline coordinator to use semantic detection in internal/pipeline/coordinator.go (check semantic patterns before basic types)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work - basic dumps with realistic data

---

## Phase 5: User Story 3 - Custom Data Patterns and Business Rules (Priority: P3)

**Goal**: Support custom data generation rules: weighted distributions, regex patterns, templates, time-series with seasonality

**Independent Test**: Specify status field with 80/15/5 distribution, generate 1000 rows, verify distribution within ±3%

### Tests for User Story 3 ⚠️

- [x] T055 [P] [US3] Unit test for weighted enum generator in tests/unit/generator/custom_test.go (TDD: distribution accuracy ±3%)
- [x] T056 [P] [US3] Unit test for pattern generator in tests/unit/generator/custom_test.go (TDD: regex pattern compliance)
- [x] T057 [P] [US3] Unit test for template generator in tests/unit/generator/custom_test.go (TDD: template placeholder replacement)
- [x] T058 [P] [US3] Unit test for timeseries generator in tests/unit/generator/timeseries_test.go (TDD: patterns, intervals, business hours)
- [x] T059 [P] [US3] Integration test for custom patterns in tests/integration/pipeline/custom_patterns_test.go (TDD: verify distributions, patterns, timeseries)

### Implementation for User Story 3

#### Custom Pattern Generators

- [x] T060 [P] [US3] Implement weighted enum generator in internal/generator/custom.go (WeightedEnumGenerator with distribution validation)
- [x] T061 [P] [US3] Implement pattern generator in internal/generator/custom.go (PatternGenerator using regex)
- [x] T062 [P] [US3] Implement template generator in internal/generator/custom.go (TemplateGenerator with year, seq, rand placeholders)
- [x] T063 [P] [US3] Implement integer range generator in internal/generator/custom.go (IntegerRangeGenerator with min/max)

#### Time-Series Generator

- [x] T064 [P] [US3] Implement time-series generator in internal/generator/timeseries.go (uniform, business_hours, daily_peak patterns)
- [x] T065 [US3] Register custom generators in internal/generator/registry.go (register weighted_enum, pattern, template, timeseries, integer_range)

#### Schema Extensions

- [x] T066 [US3] Extend schema types to support generator_config in internal/schema/types.go (add GeneratorConfig map to Column)
- [x] T067 [US3] Update schema parser to parse generator_config in internal/schema/parser.go
- [x] T068 [US3] Update pipeline to pass generator_config to generators in internal/pipeline/coordinator.go

**Checkpoint**: All user stories 1-3 should now work independently - basic, semantic, and custom data generation

---

## Phase 6: User Story 4 - Deterministic Data Generation with Seeds (Priority: P4)

**Goal**: Generate byte-identical output when same seed is used across multiple runs

**Independent Test**: Run tool twice with seed 12345, compare output files byte-for-byte (must be identical)

### Tests for User Story 4 ⚠️

- [x] T069 [P] [US4] Integration test for deterministic generation in tests/integration/pipeline/deterministic_test.go (TDD: same seed → identical output)
- [x] T070 [P] [US4] Integration test for different seeds in tests/integration/pipeline/seed_variation_test.go (TDD: different seeds → different data, same distribution)

### Implementation for User Story 4

- [x] T071 [US4] Add seed flag to generate command in internal/cli/generate.go (--seed flag, parse int64) [NOTE: Already implemented in Phase 2]
- [x] T072 [US4] Update generation context to use seed in internal/generator/context.go (initialize rand.New(rand.NewSource(seed))) [NOTE: Already implemented in Phase 2]
- [x] T073 [US4] Ensure all generators use context.Rand in internal/generator/*.go (never use global rand) [NOTE: Already implemented in US1-US3]
- [x] T074 [US4] Update pipeline to propagate seed to all workers in internal/pipeline/coordinator.go [NOTE: Already implemented in US1]

**Checkpoint**: User Stories 1-4 complete - core functionality with deterministic generation

---

## Phase 7: User Story 5 - Pre-built Scenario Templates (Priority: P5)

**Goal**: Provide pre-built templates (ecommerce, saas, healthcare, finance) that users can select and customize

**Independent Test**: Select ecommerce template, verify database has products, customers, orders, order_items, reviews, categories

### Tests for User Story 5 ⚠️

- [x] T075 [P] [US5] Unit test for template loading in tests/unit/cli/template_test.go (TDD: list, show, export templates)
- [x] T076 [P] [US5] Integration test for ecommerce template in tests/integration/templates/ecommerce_test.go (TDD: verify tables, relationships, data)
- [x] T077 [P] [US5] Integration test for saas template in tests/integration/templates/saas_test.go (TDD: verify multi-tenant structure)

### Implementation for User Story 5

#### Template Definitions

- [x] T078 [P] [US5] Create ecommerce template JSON in internal/templates/ecommerce.json (products, categories, customers, orders, order_items, reviews)
- [x] T079 [P] [US5] Create saas template JSON in internal/templates/saas.json (tenants, users, subscriptions, usage_metrics, billing)
- [x] T080 [P] [US5] Create healthcare template JSON in internal/templates/healthcare.json (patients, doctors, appointments, medical_records, prescriptions)
- [x] T081 [P] [US5] Create finance template JSON in internal/templates/finance.json (accounts, customers, transactions, investments, portfolios)

#### Template Command

- [x] T082 [US5] Implement template command in internal/cli/template.go (list, show details, export to file)
- [x] T083 [US5] Embed templates using go:embed in internal/templates/embed.go
- [x] T084 [US5] Add template loading to generate command in internal/cli/generate.go (--template flag, --template-param flag)
- [x] T085 [US5] Implement template parameter substitution in internal/templates/embed.go (override row counts, customize values)
- [x] T086 [US5] Wire template command to root in internal/cli/root.go

**Checkpoint**: User Stories 1-5 complete - full functionality with templates

---

## Phase 8: User Story 6 - Multiple Output Format Support (Priority: P6)

**Goal**: Support pg_dump custom format (default), SQL INSERT format, and COPY format

**Independent Test**: Generate same schema in all three formats, verify each can be imported to PostgreSQL

### Tests for User Story 6 ⚠️

- [x] T087 [P] [US6] Unit test for SQL format writer in tests/unit/pgdump/sql_writer_test.go (TDD: INSERT statements, batching)
- [x] T088 [P] [US6] Unit test for COPY format writer in tests/unit/pgdump/copy_writer_test.go (TDD: COPY commands, TSV data)
- [ ] T089 [P] [US6] Integration test for SQL format in tests/integration/postgresql/sql_format_test.go (TDD: restore via psql)
- [ ] T090 [P] [US6] Integration test for COPY format in tests/integration/postgresql/copy_format_test.go (TDD: restore via psql)

### Implementation for User Story 6

#### SQL Format Writer

- [x] T091 [P] [US6] Implement SQL format writer in internal/pgdump/sql_writer.go (CREATE TABLE, INSERT statements, batched inserts)
- [x] T092 [P] [US6] Implement SQL escaping and quoting in internal/pgdump/sql_escape.go (escape strings, identifiers, NULL values)

#### COPY Format Writer

- [x] T093 [P] [US6] Implement COPY format writer in internal/pgdump/copy_writer.go (CREATE TABLE, COPY FROM stdin, TSV data)
- [x] T094 [P] [US6] Implement COPY data escaping in internal/pgdump/copy_escape.go (tab, newline, backslash escaping)

#### Format Selection

- [x] T095 [US6] Add format flag to generate command in internal/cli/generate.go (--format: custom|sql|copy)
- [x] T096 [US6] Update pipeline coordinator to select writer based on format in internal/pipeline/coordinator.go
- [x] T097 [US6] Update dump writer factory in internal/pgdump/writer.go (factory pattern for format selection)

**Checkpoint**: All user stories complete and independently functional

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

### Validation & Error Handling

- [ ] T098 [P] Write unit test for validate command in tests/unit/cli/validate_test.go (TDD: validation output, JSON format)
- [ ] T099 [P] Implement validate command in internal/cli/validate.go (validate schema without generating data)
- [ ] T100 [P] Add detailed error messages to schema validator in internal/schema/validator.go (line numbers, suggestions)
- [ ] T101 [P] Add pg_query validation option in internal/pgdump/validate.go (optional --validate-output flag)

### Performance & Concurrency

- [ ] T102 [P] Write unit test for worker pool in tests/unit/pipeline/workers_test.go (TDD: concurrent generation, backpressure)
- [ ] T103 [P] Implement worker pool in internal/pipeline/workers.go (configurable workers, channel coordination)
- [ ] T104 [P] Write unit test for LRU cache in tests/unit/generator/cache_test.go (TDD: FK lookups, eviction)
- [ ] T105 [P] Implement LRU cache for FK lookups in internal/generator/cache.go (hashicorp/golang-lru v2)
- [ ] T106 [P] Add --jobs flag to generate command in internal/cli/generate.go (control worker count)
- [ ] T107 [P] Implement streaming write for large tables in internal/pgdump/writer.go (batch size 1000 rows)

### Configuration & Logging

- [ ] T108 [P] Implement config file loading in internal/cli/config.go (Viper, .datagen.yaml, precedence)
- [ ] T109 [P] Add structured logging in internal/cli/logging.go (zerolog, security events, --verbose)
- [ ] T110 [P] Add progress indicators to generate command in internal/cli/progress.go (progress bars, table completion)

### Documentation & Examples

- [ ] T111 [P] Create example schemas in docs/examples/ (blog, ecommerce-custom, analytics)
- [ ] T112 [P] Generate man pages in docs/man/ (using Cobra's GenManTree)
- [ ] T113 [P] Create architecture documentation in docs/architecture.md
- [ ] T114 [P] Create generator documentation in docs/generators.md (list all generators, examples)

### Build & Release

- [ ] T115 [P] Create build script in scripts/build.sh (cross-compile for linux/darwin/windows, amd64/arm64)
- [ ] T116 [P] Create test script in scripts/test.sh (unit, integration, e2e, coverage report)
- [ ] T117 [P] Create release script in scripts/release.sh (GitHub releases, checksums, changelog)
- [ ] T118 [P] Setup Homebrew tap configuration in homebrew/datagen.rb
- [ ] T119 [P] Create Dockerfile for Alpine-based image in Dockerfile
- [ ] T120 [P] Add performance benchmarks to tests/benchmarks/ (benchmark all generators, pipeline throughput)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4 → P5 → P6)
- **Polish (Phase 9)**: Depends on desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Builds on US1 pipeline but independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Extends generators from US1/US2 but independently testable
- **User Story 4 (P4)**: Can start after Foundational (Phase 2) - Modifies seeding in US1 but independently testable
- **User Story 5 (P5)**: Can start after Foundational (Phase 2) - Uses US1-US3 generators but independently testable
- **User Story 6 (P6)**: Can start after US1 complete (needs dump writer abstraction) - Independently testable

### Within Each User Story

**TDD Workflow (Constitution Requirement)**:
1. Tests MUST be written FIRST and MUST FAIL initially (Red phase)
2. Implementation proceeds only after test failure verified (Green phase)
3. Refactoring after tests pass (Refactor phase)

**Task Order within Story**:
- Tests before implementation (TDD)
- Models/types before services/generators
- Generators before pipeline integration
- Core implementation before CLI integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2 groups)
- Once Foundational phase completes, user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can be written in parallel
- Implementation tasks marked [P] within a story can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Write all tests for User Story 1 together (TDD - Red phase):
T022: "Contract test for basic type generators in tests/unit/generator/basic_test.go"
T023: "Integration test for simple schema pipeline in tests/integration/pipeline/basic_pipeline_test.go"
T024: "Integration test for PostgreSQL restore in tests/integration/postgresql/restore_test.go"

# Verify tests FAIL (no implementation yet)

# Implement parallelizable components (Green phase):
T025: "Implement basic type generators in internal/generator/basic.go"
T026: "Implement sequence generator in internal/generator/sequence.go"
T028: "Write unit test for dump header in tests/unit/pgdump/header_test.go"
T029: "Implement dump header writer in internal/pgdump/header.go"
T030: "Write unit test for TOC entry in tests/unit/pgdump/toc_test.go"
T031: "Implement TOC structure in internal/pgdump/toc.go"
T034: "Write unit test for compression in tests/unit/pgdump/compression_test.go"
T035: "Implement gzip compression in internal/pgdump/compression.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
   - Create simple 2-table schema (users, posts)
   - Generate dump file
   - Restore to PostgreSQL with pg_restore
   - Verify tables, data, constraints
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo (realistic data)
4. Add User Story 3 → Test independently → Deploy/Demo (custom patterns)
5. Add User Story 4 → Test independently → Deploy/Demo (deterministic)
6. Add User Story 5 → Test independently → Deploy/Demo (templates)
7. Add User Story 6 → Test independently → Deploy/Demo (multiple formats)
8. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (MVP critical path)
   - Developer B: User Story 2 (semantic generators)
   - Developer C: User Story 3 (custom patterns)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- **TDD MANDATORY**: Tests MUST be written BEFORE implementation (constitution requirement)
- Verify tests fail before implementing (Red phase)
- Implement to make tests pass (Green phase)
- Refactor after tests pass (Refactor phase)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Test coverage MUST be ≥90% for new code
- Integration tests MUST verify PostgreSQL archive format compliance
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence

---

## Test Coverage Requirements (Constitution)

Per constitution, test coverage MUST be ≥90% for new code:

- **Unit Tests (70% of test effort)**: All generators, parsers, validators, writers
- **Integration Tests (25% of test effort)**: Pipeline flow, PostgreSQL restore validation
- **End-to-End Tests (5% of test effort)**: Full CLI workflows with real schemas

**Test Types Required**:
1. Unit tests for each generator type
2. Contract tests for all public interfaces
3. Integration tests with real PostgreSQL via testcontainers-go
4. Property-based tests for data distributions
5. Golden file regression tests for SQL output
6. Performance benchmarks for generation speed

**PostgreSQL Version Coverage**:
- Integration tests MUST verify compatibility with PostgreSQL 12, 13, 14, 15, 16
- Use testcontainers-go to spin up each version
- Verify pg_restore succeeds for all versions