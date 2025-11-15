package schema

// Schema represents the top-level database schema definition
type Schema struct {
	Version     string                 `json:"version"`
	Database    DatabaseConfig         `json:"database"`
	Tables      map[string]*Table      `json:"tables"`
	Sequences   map[string]*Sequence   `json:"sequences,omitempty"`
	CustomTypes map[string]*CustomType `json:"custom_types,omitempty"`
	Extensions  []string               `json:"extensions,omitempty"`
}

// DatabaseConfig represents database-level configuration
type DatabaseConfig struct {
	Name     string `json:"name"`
	Encoding string `json:"encoding"` // default: "UTF8"
	Locale   string `json:"locale"`   // default: "en_US.utf8"
}

// Table represents a single database table
type Table struct {
	Columns           []*Column           `json:"columns"`
	PrimaryKey        []string            `json:"primary_key,omitempty"`
	ForeignKeys       []*ForeignKey       `json:"foreign_keys,omitempty"`
	UniqueConstraints []*UniqueConstraint `json:"unique_constraints,omitempty"`
	CheckConstraints  []*CheckConstraint  `json:"check_constraints,omitempty"`
	Indexes           []*Index            `json:"indexes,omitempty"`
	RowCount          int                 `json:"row_count"`

	// Computed fields (not in JSON)
	Dependencies []string `json:"-"`
}

// Column represents a table column
type Column struct {
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Nullable        bool                   `json:"nullable,omitempty"`
	DefaultValue    string                 `json:"default,omitempty"`
	Unique          bool                   `json:"unique,omitempty"`
	PrimaryKey      bool                   `json:"primary_key,omitempty"`
	GeneratorType   string                 `json:"generator,omitempty"`
	GeneratorConfig map[string]interface{} `json:"generator_config,omitempty"`
	Comment         string                 `json:"comment,omitempty"`
}

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	Columns           []string `json:"columns"`
	ReferencedTable   string   `json:"referenced_table"`
	ReferencedColumns []string `json:"referenced_columns"`
	OnDelete          string   `json:"on_delete,omitempty"` // CASCADE, SET NULL, RESTRICT, NO ACTION
	OnUpdate          string   `json:"on_update,omitempty"`
}

// UniqueConstraint represents a unique constraint
type UniqueConstraint struct {
	Columns []string `json:"columns"`
	Name    string   `json:"name,omitempty"`
}

// CheckConstraint represents a check constraint
type CheckConstraint struct {
	Expression string `json:"expression"`
	Name       string `json:"name,omitempty"`
}

// Index represents a database index
type Index struct {
	Columns []string `json:"columns"`
	Name    string   `json:"name,omitempty"`
	Type    string   `json:"type,omitempty"` // btree, hash, gin, gist
	Unique  bool     `json:"unique,omitempty"`
}

// Sequence represents a PostgreSQL sequence
type Sequence struct {
	Start     int64  `json:"start"`
	Increment int64  `json:"increment"`
	MinValue  *int64 `json:"min_value,omitempty"`
	MaxValue  *int64 `json:"max_value,omitempty"`
	Cache     int64  `json:"cache,omitempty"`
	Cycle     bool   `json:"cycle,omitempty"`
}

// CustomType represents a user-defined PostgreSQL type
type CustomType struct {
	Kind       string      `json:"kind"` // enum, composite, domain
	Definition interface{} `json:"definition"`
}

// EnumDefinition represents an enum type definition
type EnumDefinition struct {
	Values []string `json:"values"`
}

// CompositeDefinition represents a composite type definition
type CompositeDefinition struct {
	Fields []Column `json:"fields"`
}

// DomainDefinition represents a domain type definition
type DomainDefinition struct {
	BaseType   string `json:"base_type"`
	Constraint string `json:"constraint,omitempty"`
}