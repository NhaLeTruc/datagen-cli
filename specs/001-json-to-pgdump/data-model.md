# Data Model: JSON Schema to PostgreSQL Dump Generator

**Feature**: 001-json-to-pgdump
**Created**: 2025-11-15
**Purpose**: Define internal domain models and their relationships for the datagen-cli tool

## Overview

This document defines the core domain entities used internally by the datagen-cli tool. These models represent the transformation pipeline from JSON input to PostgreSQL dump output.

## Entity Diagram

```
┌─────────────┐
│   Schema    │
└──────┬──────┘
       │ 1
       │ has
       │ *
┌──────▼──────┐      ┌───────────────┐
│    Table    │      │  CustomType   │
└──────┬──────┘      └───────────────┘
       │ 1
       │ has
       │ *
┌──────▼──────┐
│   Column    │
└──────┬──────┘
       │ 1
       │ uses
       │ 1
┌──────▼──────────┐
│    Generator    │
└─────────────────┘

┌────────────┐
│  Template  │──── uses ───▶ Schema
└────────────┘

┌────────────┐
│  DumpFile  │──── contains ───▶ Schema
└────────────┘
```

## Core Entities

### Schema

**Purpose**: Top-level representation of a complete PostgreSQL database schema

**Fields**:
- `Version` (string): Schema format version (e.g., "1.0")
- `Database` (DatabaseConfig): Database-level configuration
- `Tables` (map[string]*Table): Tables indexed by name
- `Sequences` (map[string]*Sequence): Sequences indexed by name
- `CustomTypes` (map[string]*CustomType): User-defined types indexed by name
- `Extensions` ([]string): PostgreSQL extensions to enable (e.g., "uuid-ossp", "pgcrypto")

**Relationships**:
- Has many Tables
- Has many Sequences
- Has many CustomTypes

**Validation Rules**:
- Version must match supported format versions (currently "1.0")
- Database name must be valid PostgreSQL identifier (alphanumeric + underscore, starts with letter)
- Table names must be unique within schema
- Sequence names must be unique within schema
- CustomType names must be unique within schema
- No circular table dependencies for foreign keys

**State Transitions**:
```
Created → Parsed → Validated → DependenciesResolved → Ready
```

**Go Type Signature**:
```go
type Schema struct {
    Version     string                 `json:"version"`
    Database    DatabaseConfig         `json:"database"`
    Tables      map[string]*Table      `json:"tables"`
    Sequences   map[string]*Sequence   `json:"sequences,omitempty"`
    CustomTypes map[string]*CustomType `json:"custom_types,omitempty"`
    Extensions  []string               `json:"extensions,omitempty"`
}

type DatabaseConfig struct {
    Name     string `json:"name"`
    Encoding string `json:"encoding"` // default: "UTF8"
    Locale   string `json:"locale"`   // default: "en_US.utf8"
}
```

---

### Table

**Purpose**: Represents a single PostgreSQL table with its structure and data generation configuration

**Fields**:
- `Name` (string): Table name
- `Columns` ([]*Column): List of column definitions
- `PrimaryKey` ([]string): Column names comprising primary key
- `ForeignKeys` ([]*ForeignKey): Foreign key constraints
- `UniqueConstraints` ([]*UniqueConstraint): Unique constraints
- `CheckConstraints` ([]*CheckConstraint): Check constraints
- `Indexes` ([]*Index): Index definitions
- `RowCount` (int): Number of rows to generate
- `Dependencies` ([]string): Table names this table depends on (computed from foreign keys)

**Relationships**:
- Belongs to Schema
- Has many Columns
- Has many ForeignKeys (references other Tables)
- Has many Indexes

**Validation Rules**:
- Name must be valid PostgreSQL identifier
- Must have at least one column
- PrimaryKey columns must exist in Columns
- ForeignKey referenced tables must exist in schema
- ForeignKey referenced columns must exist and match types
- UniqueConstraint columns must exist
- CheckConstraint expressions must be valid SQL
- RowCount must be ≥ 0 (0 means empty table)
- Circular foreign key dependencies must be detected and handled

**State Transitions**:
```
Created → ColumnsAdded → ConstraintsAdded → Validated → Ready
```

**Go Type Signature**:
```go
type Table struct {
    Name              string              `json:"-"`
    Columns           []*Column           `json:"columns"`
    PrimaryKey        []string            `json:"primary_key,omitempty"`
    ForeignKeys       []*ForeignKey       `json:"foreign_keys,omitempty"`
    UniqueConstraints []*UniqueConstraint `json:"unique_constraints,omitempty"`
    CheckConstraints  []*CheckConstraint  `json:"check_constraints,omitempty"`
    Indexes           []*Index            `json:"indexes,omitempty"`
    RowCount          int                 `json:"row_count"`

    // Computed fields
    Dependencies []string `json:"-"`
}

type ForeignKey struct {
    Columns          []string `json:"columns"`
    ReferencedTable  string   `json:"referenced_table"`
    ReferencedColumns []string `json:"referenced_columns"`
    OnDelete         string   `json:"on_delete,omitempty"` // CASCADE, SET NULL, RESTRICT, NO ACTION
    OnUpdate         string   `json:"on_update,omitempty"`
}

type UniqueConstraint struct {
    Columns []string `json:"columns"`
    Name    string   `json:"name,omitempty"`
}

type CheckConstraint struct {
    Expression string `json:"expression"`
    Name       string `json:"name,omitempty"`
}

type Index struct {
    Columns []string `json:"columns"`
    Name    string   `json:"name,omitempty"`
    Type    string   `json:"type,omitempty"` // btree, hash, gin, gist
    Unique  bool     `json:"unique,omitempty"`
}
```

---

### Column

**Purpose**: Defines a single column in a table, including its type, constraints, and data generation strategy

**Fields**:
- `Name` (string): Column name
- `Type` (string): PostgreSQL data type (e.g., "varchar(255)", "integer", "timestamp", "uuid")
- `Nullable` (bool): Whether column allows NULL values
- `DefaultValue` (string): Default value expression (SQL)
- `GeneratorType` (string): Generator to use for data (e.g., "email", "phone", "uuid", "sequential")
- `GeneratorConfig` (map[string]interface{}): Additional configuration for generator
- `Comment` (string): Column comment for documentation

**Relationships**:
- Belongs to Table
- Uses one Generator (determined by GeneratorType)

**Validation Rules**:
- Name must be valid PostgreSQL identifier
- Type must be valid PostgreSQL data type
- If Nullable is false, DefaultValue or GeneratorType must be specified
- GeneratorType must exist in GeneratorRegistry
- GeneratorConfig must match generator's expected schema
- For primary key columns, Nullable must be false

**Data Generation Logic**:
1. If column is primary key and Type is serial/bigserial, use sequence generator
2. Else if GeneratorType is explicitly specified, use that generator
3. Else if column name matches semantic pattern (email, phone, etc.), use semantic generator
4. Else use type-based default generator (e.g., random integer for integer type)
5. Apply GeneratorConfig to customize behavior

**Go Type Signature**:
```go
type Column struct {
    Name            string                 `json:"name"`
    Type            string                 `json:"type"`
    Nullable        bool                   `json:"nullable,omitempty"`
    DefaultValue    string                 `json:"default,omitempty"`
    GeneratorType   string                 `json:"generator,omitempty"`
    GeneratorConfig map[string]interface{} `json:"generator_config,omitempty"`
    Comment         string                 `json:"comment,omitempty"`
}
```

---

### Generator

**Purpose**: Interface and implementations for generating realistic data for columns

**Interface**:
```go
type Generator interface {
    // Generate produces a single value for the column
    Generate(ctx *GenerationContext) (interface{}, error)

    // Validate checks if a value is valid for this generator
    Validate(value interface{}) error

    // Type returns the PostgreSQL type this generator produces
    Type() string
}

type GenerationContext struct {
    Rand       *rand.Rand           // Seeded random for deterministic generation
    RowIndex   int                  // Current row being generated (0-based)
    TableName  string               // Table being generated
    ColumnName string               // Column being generated
    FKCache    *LRUCache            // Cache of foreign key values for lookups
    Config     map[string]interface{} // Generator-specific configuration
}
```

**Generator Types**:

1. **SemanticGenerator**: Detects column names and generates appropriate data
   - Email: "email", "email_address", "user_email"
   - Phone: "phone", "phone_number", "mobile", "tel"
   - Name: "first_name", "last_name", "full_name", "name"
   - Address: "address", "street", "city", "state", "country", "zip", "postal_code"
   - Date/Time: "created_at", "updated_at", "birthdate", "timestamp"
   - UUID: "id", "uuid", "guid" (if type is uuid)

2. **BasicTypeGenerator**: Generates data based on PostgreSQL type
   - Integer types: Random integers within type bounds
   - Text types: Random strings of appropriate length
   - Boolean: Random true/false
   - Timestamp: Random recent timestamps
   - Numeric: Random decimal numbers
   - UUID: Random UUIDs

3. **CustomPatternGenerator**: User-defined patterns
   - Regex-based: Generate strings matching regex
   - Template-based: Fill template with placeholders (e.g., "USER-{year}-{seq}")
   - Enum-based: Select from predefined list of values
   - Distribution-based: Weighted random selection

4. **TimeSeriesGenerator**: Generate time-series data
   - Sequential timestamps with configurable interval
   - Daily/weekly/monthly patterns
   - Business hours filtering
   - Seasonality and trends

5. **ForeignKeyGenerator**: Generate foreign key values
   - Cache referenced table's primary keys
   - Random selection from cache
   - Ensure referential integrity

6. **SequenceGenerator**: Generate sequential numbers
   - Auto-increment for serial types
   - Configurable start and increment values

**Registry Pattern**:
```go
type GeneratorRegistry struct {
    generators map[string]GeneratorFactory
    mu         sync.RWMutex
}

type GeneratorFactory func(config map[string]interface{}) (Generator, error)

// Register a generator
func (r *GeneratorRegistry) Register(name string, factory GeneratorFactory)

// Get a generator by name
func (r *GeneratorRegistry) Get(name string, config map[string]interface{}) (Generator, error)
```

---

### Sequence

**Purpose**: Represents a PostgreSQL sequence object

**Fields**:
- `Name` (string): Sequence name
- `Start` (int64): Starting value
- `Increment` (int64): Increment value
- `MinValue` (*int64): Minimum value (nil = no minimum)
- `MaxValue` (*int64): Maximum value (nil = no maximum)
- `Cache` (int64): Cache size for performance
- `Cycle` (bool): Whether to cycle when reaching max

**Go Type Signature**:
```go
type Sequence struct {
    Name      string `json:"name"`
    Start     int64  `json:"start"`
    Increment int64  `json:"increment"`
    MinValue  *int64 `json:"min_value,omitempty"`
    MaxValue  *int64 `json:"max_value,omitempty"`
    Cache     int64  `json:"cache,omitempty"`
    Cycle     bool   `json:"cycle,omitempty"`
}
```

---

### CustomType

**Purpose**: Represents a user-defined PostgreSQL type (enum, composite, domain)

**Fields**:
- `Name` (string): Type name
- `Kind` (string): Type kind ("enum", "composite", "domain")
- `Definition` (interface{}): Type-specific definition

**Go Type Signature**:
```go
type CustomType struct {
    Name       string      `json:"name"`
    Kind       string      `json:"kind"` // enum, composite, domain
    Definition interface{} `json:"definition"`
}

// For enums
type EnumDefinition struct {
    Values []string `json:"values"`
}

// For composite types
type CompositeDefinition struct {
    Fields []Column `json:"fields"`
}

// For domains
type DomainDefinition struct {
    BaseType   string   `json:"base_type"`
    Constraint string   `json:"constraint,omitempty"`
}
```

---

### Template

**Purpose**: Pre-built schema definition for common use cases

**Fields**:
- `Name` (string): Template name (e.g., "ecommerce", "saas")
- `Description` (string): Template description
- `Category` (string): Template category
- `Schema` (*Schema): The schema definition
- `Parameters` (map[string]TemplateParameter): Customizable parameters

**Customization**:
Templates can be customized by overriding parameters:
- `row_counts`: Override row counts per table
- `enabled_tables`: Select subset of tables to include
- `locale`: Change data generation locale

**Go Type Signature**:
```go
type Template struct {
    Name        string                        `json:"name"`
    Description string                        `json:"description"`
    Category    string                        `json:"category"`
    Schema      *Schema                       `json:"schema"`
    Parameters  map[string]TemplateParameter  `json:"parameters,omitempty"`
}

type TemplateParameter struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"` // int, string, bool, string_array
    Default     interface{} `json:"default"`
    Description string      `json:"description"`
}
```

---

### DumpFile

**Purpose**: Represents the output PostgreSQL dump file structure

**Fields**:
- `Format` (string): Output format ("custom", "sql", "copy")
- `Header` (*DumpHeader): File header with metadata
- `TOC` ([]*TOCEntry): Table of contents
- `DataSections` ([]*DataSection): Data blocks
- `Compression` (string): Compression algorithm ("gzip", "none")

**Structure (Custom Format)**:
```
[Header: PGDMP magic bytes + version + metadata]
[TOC: Linked list of entries with dependencies]
[Data Section 1: Schema definitions]
[Data Section 2: Table data]
[Data Section 3: Indexes]
[Data Section 4: Constraints]
...
```

**Go Type Signature**:
```go
type DumpFile struct {
    Format       string          `json:"-"`
    Header       *DumpHeader     `json:"-"`
    TOC          []*TOCEntry     `json:"-"`
    DataSections []*DataSection  `json:"-"`
    Compression  string          `json:"-"`
}

type DumpHeader struct {
    MagicBytes    [5]byte // "PGDMP"
    VersionMajor  int
    VersionMinor  int
    VersionRev    int
    IntSize       int
    OffSize       int
    FormatVersion int
    Compression   int
    Timestamp     time.Time
    DatabaseName  string
}

type TOCEntry struct {
    ID           int
    DataOID      int
    Description  string
    Namespace    string
    Tag          string
    Owner        string
    Dependencies []int
    DataOffset   int64
    DataLength   int64
}

type DataSection struct {
    TOCID   int
    Data    []byte
    Compressed bool
}
```

---

## Data Flow

### Schema Parsing Flow
```
JSON Input → Parser → Schema struct → Validator → DependencyResolver → Ready
```

### Data Generation Flow
```
Schema → Table (sorted by dependencies) → Column → Generator → Value → DumpWriter
```

### Pipeline Execution
```
1. Parse JSON schema → Schema struct
2. Validate schema (types, constraints, references)
3. Resolve table dependencies → Topological sort
4. Initialize generators for each column
5. For each table (in dependency order):
   a. Create worker pool
   b. Generate rows in batches
   c. Write to dump file (streaming)
   d. Cache primary keys for foreign key references
6. Write indexes and constraints
7. Finalize dump file (TOC, header)
```

## Validation Rules Summary

| Entity | Key Validations |
|--------|----------------|
| Schema | Version valid, unique table/sequence/type names, no circular FK dependencies |
| Table | Valid identifier, ≥1 column, PK columns exist, FK references valid, row count ≥0 |
| Column | Valid identifier, valid PG type, nullable/default consistency, generator exists |
| Generator | Produces values matching column type, config matches schema |
| Sequence | Start < Max (if set), Increment ≠ 0 |
| CustomType | Valid kind, definition matches kind requirements |
| Template | Valid schema, parameters have valid types and defaults |
| DumpFile | Format is supported, header valid, TOC entries reference valid data sections |

## Performance Considerations

### Memory Management
- **Tables**: Process one at a time in dependency order
- **Rows**: Generate and write in batches (configurable batch size, default 1000)
- **Foreign Keys**: LRU cache for referenced values (size: min(10000, referenced_table_rows))
- **Generators**: Use sync.Pool for frequently allocated objects
- **Dump File**: Streaming write, never hold full dump in memory

### Concurrency
- **Table-level parallelism**: Independent tables (no FK dependencies) generated concurrently
- **Row-level parallelism**: Worker pool per table (workers = min(CPU cores, 4))
- **Generator thread-safety**: Each worker has its own seeded Rand instance
- **Cache thread-safety**: LRU cache uses RWMutex for concurrent access

### Caching Strategy
- **Foreign Key Cache**: LRU with size = min(10000, referenced_rows)
- **Type Metadata Cache**: In-memory map (small, <1MB)
- **Generator Registry**: Singleton with lazy initialization
- **Template Cache**: Embed templates in binary, parse once on first use