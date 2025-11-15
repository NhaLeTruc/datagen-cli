<!--
Sync Impact Report:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Version Change: N/A → 1.0.0
Change Type: MAJOR - Initial constitution ratification

Modified/Added Principles:
  ✓ I. Test-First Development (NON-NEGOTIABLE) - Added
  ✓ II. Clean Code Standards - Added
  ✓ III. Security-First Architecture - Added
  ✓ IV. PostgreSQL Archive Integrity - Added
  ✓ V. CLI Interface Contract - Added
  ✓ VI. Performance & Scale - Added

Added Sections:
  ✓ Security Requirements
  ✓ Development Workflow
  ✓ Governance

Templates Requiring Updates:
  ✅ .specify/templates/plan-template.md - Reviewed (already supports constitution checks)
  ✅ .specify/templates/spec-template.md - Reviewed (compatible with principles)
  ✅ .specify/templates/tasks-template.md - Reviewed (supports TDD workflow with test-first ordering)

Follow-up TODOs: None

Notes:
  - Constitution emphasizes TDD as non-negotiable requirement
  - Security principles include input validation, secure file handling, and audit logging
  - PostgreSQL-specific principles ensure archive format correctness and data integrity
  - Clean code standards enforce SOLID principles and code quality metrics

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
-->

# DataGen-CLI Constitution

## Core Principles

### I. Test-First Development (NON-NEGOTIABLE)

**TDD is mandatory for ALL code changes without exception.**

- Tests MUST be written BEFORE implementation code
- Tests MUST fail initially (Red phase)
- Implementation proceeds only after test failure is verified (Green phase)
- Refactoring only after tests pass (Refactor phase)
- Red-Green-Refactor cycle is strictly enforced
- No pull request shall be accepted without corresponding tests
- Test coverage MUST be maintained at ≥90% for new code
- Integration tests MUST verify PostgreSQL archive format compliance

**Rationale**: TDD ensures correctness by design, prevents regressions, and serves as living documentation. For a tool generating database archives, correctness is non-negotiable—corrupted archives can cause catastrophic data loss.

### II. Clean Code Standards

**All code MUST adhere to clean code principles and measurable quality metrics.**

- **SOLID Principles**: Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion MUST be followed
- **Naming**: Variables, functions, classes use descriptive, unambiguous names (no single-letter variables except loop counters)
- **Function Size**: Maximum 20 lines per function; functions doing >1 thing MUST be split
- **Cyclomatic Complexity**: Maximum complexity of 10 per function
- **Code Duplication**: Zero tolerance for copy-paste code; extract to shared functions
- **Comments**: Code MUST be self-documenting; comments only for "why", not "what"
- **Error Handling**: All error paths explicitly handled; no silent failures
- **Type Safety**: Strong typing enforced; no `any` types or equivalents without justification

**Rationale**: Clean code reduces bugs, accelerates onboarding, and ensures long-term maintainability. Archive generation involves complex binary formats—clean abstractions prevent subtle corruption bugs.

### III. Security-First Architecture

**Security MUST be embedded in every layer of the application.**

- **Input Validation**: ALL user inputs (CLI args, config files, templates) validated against strict schemas
- **Path Traversal Prevention**: File paths sanitized; operations restricted to designated directories
- **Injection Prevention**: SQL templates parameterized; no string concatenation for queries
- **Least Privilege**: File operations use minimal required permissions
- **Secure Defaults**: All features default to secure configurations
- **Secrets Management**: No credentials, tokens, or keys in code or version control
- **Dependency Scanning**: Automated vulnerability scanning for all dependencies
- **Audit Logging**: Security-relevant operations (file access, archive generation) logged with timestamps and user context

**Rationale**: CLI tools often run with elevated privileges. Archive files may contain sensitive data. A security breach could expose confidential database schemas or data patterns.

### IV. PostgreSQL Archive Integrity

**Generated archives MUST be 100% compatible with PostgreSQL tooling.**

- **Format Compliance**: Archives MUST match PostgreSQL archive format specification (pg_dump custom format)
- **Version Testing**: Archives tested against PostgreSQL versions 12, 13, 14, 15, 16
- **Integrity Verification**: Every generated archive MUST be restorable via `pg_restore`
- **Metadata Accuracy**: Archive headers, table of contents, and dependencies MUST be valid
- **Compression Support**: Support for gzip and custom compression formats
- **Large Object Handling**: Proper encoding of BLOBs and large text fields
- **Character Encoding**: UTF-8 support with proper escaping
- **Transaction Consistency**: Archives MUST represent consistent snapshots

**Rationale**: Incompatible archives render the tool useless. Users depend on archives for backup testing, data migration, and development workflows. Format errors can cause data loss or corruption.

### V. CLI Interface Contract

**CLI MUST provide predictable, composable, and user-friendly interface.**

- **Text Protocol**: Input via stdin/arguments, output to stdout, errors to stderr
- **Exit Codes**: 0 for success, non-zero for failure (standard POSIX conventions)
- **JSON & Human Formats**: All outputs support both `--json` and human-readable formats
- **Idempotency**: Same inputs MUST produce identical outputs (deterministic generation)
- **Progress Reporting**: Long operations provide progress indicators (spinners, percentages)
- **Configuration Files**: Support for JSON/YAML config files to avoid long command lines
- **Help & Documentation**: `--help` provides complete usage; examples in docs
- **Backwards Compatibility**: CLI flags and behavior versioned; breaking changes require major version bump

**Rationale**: CLI tools are often scripted. Predictable interfaces enable automation, CI/CD integration, and composability with other tools.

### VI. Performance & Scale

**Tool MUST handle production-scale workloads efficiently.**

- **Memory Efficiency**: Streaming architecture; no loading entire datasets in memory
- **Target Performance**: Generate 1GB archive in <30 seconds on standard hardware
- **Scalability**: Support for archives up to 100GB
- **Resource Limits**: Configurable memory and CPU limits
- **Parallel Generation**: Support for parallelizing data generation across tables
- **Benchmarking**: Performance regression tests in CI/CD pipeline
- **Profiling**: Regular profiling to identify bottlenecks

**Rationale**: Real-world testing scenarios require large datasets. Slow tools impede development velocity. Memory-efficient design enables usage on constrained environments.

## Security Requirements

### Threat Model

**Identified threats this tool MUST defend against:**

1. **Malicious Templates**: User-supplied templates executing arbitrary code
2. **Path Traversal**: Archive operations accessing unauthorized filesystem locations
3. **Resource Exhaustion**: Malicious inputs causing DoS via memory/CPU exhaustion
4. **Information Disclosure**: Stack traces or errors leaking sensitive system information
5. **Dependency Vulnerabilities**: Compromised or vulnerable third-party libraries

### Security Controls

**Mandatory security implementations:**

- **Template Sandboxing**: Template engines run in restricted execution context
- **Input Sanitization**: Whitelist-based validation for all file paths and SQL identifiers
- **Rate Limiting**: Configurable limits on archive size, table count, row count
- **Secure Error Handling**: Production mode hides implementation details
- **SBOM Generation**: Software Bill of Materials for all dependencies
- **Automated Updates**: Dependabot or equivalent for dependency patching
- **Security Audits**: Quarterly manual security reviews

## Development Workflow

### Code Review Requirements

**All code changes MUST pass peer review:**

- Minimum 1 approver required (2 for security-sensitive changes)
- Reviewer MUST verify test-first workflow evidence
- Reviewer MUST confirm tests cover edge cases
- Reviewer MUST check for OWASP Top 10 vulnerabilities
- Automated checks: linting, formatting, test coverage, security scanning

### Quality Gates

**CI/CD pipeline MUST enforce these gates:**

1. **Linting**: Code passes all linter rules (zero warnings)
2. **Formatting**: Code matches automated formatter output
3. **Type Checking**: All type checks pass
4. **Unit Tests**: 100% of unit tests pass; coverage ≥90%
5. **Integration Tests**: All PostgreSQL version tests pass
6. **Security Scan**: Zero high/critical vulnerabilities
7. **Performance Tests**: No regressions >5% from baseline
8. **License Compliance**: All dependencies use approved licenses

### Testing Discipline

**Test pyramid structure:**

- **Unit Tests**: 70% - Fast, isolated, test individual functions
- **Integration Tests**: 25% - Verify PostgreSQL archive compatibility
- **End-to-End Tests**: 5% - Full CLI workflows

**Test naming convention**: `test_<method>_<scenario>_<expected_outcome>`

**Test data**: Fixtures managed in version control; no hardcoded test data in tests

## Governance

### Amendment Process

**Constitution amendments require:**

1. Written proposal with rationale
2. Impact analysis on existing code and workflows
3. Approval from majority of maintainers
4. Migration plan for non-compliant code
5. Version bump following semantic versioning

### Versioning Policy

- **MAJOR**: Backward-incompatible principle removal or redefinition
- **MINOR**: New principle added or existing principle materially expanded
- **PATCH**: Clarifications, wording improvements, typo fixes

### Compliance Reviews

**Quarterly compliance audits MUST verify:**

- All merged code follows TDD workflow (evidenced by commit history)
- Test coverage meets thresholds
- Security scans show no critical issues
- Performance benchmarks within acceptable ranges
- Documentation is up-to-date

### Complexity Justification

**Any deviation from constitutional principles MUST:**

- Be documented in `plan.md` Complexity Tracking section
- Explain why simpler alternatives were rejected
- Receive explicit approval from tech lead
- Include remediation plan if temporary

**Version**: 1.0.0 | **Ratified**: 2025-11-15 | **Last Amended**: 2025-11-15