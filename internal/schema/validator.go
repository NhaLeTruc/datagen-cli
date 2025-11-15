package schema

import (
	"fmt"
	"strings"
)

// Validate checks the schema for common errors and returns a list of validation errors
func Validate(s *Schema) []error {
	var errs []error

	// Validate each table
	for tableName, table := range s.Tables {
		errs = append(errs, validateTable(tableName, table, s)...)
	}

	// Check for circular dependencies
	errs = append(errs, validateDependencies(s)...)

	return errs
}

func validateTable(name string, t *Table, s *Schema) []error {
	var errs []error

	// Validate row count
	if t.RowCount <= 0 {
		errs = append(errs, fmt.Errorf("table %s: row_count must be greater than 0, got %d", name, t.RowCount))
	}

	// Validate columns
	if len(t.Columns) == 0 {
		errs = append(errs, fmt.Errorf("table %s: must have at least one column", name))
		return errs
	}

	columnNames := make(map[string]bool)
	for _, col := range t.Columns {
		errs = append(errs, validateColumn(name, col)...)
		columnNames[col.Name] = true
	}

	// Validate foreign keys
	for _, fk := range t.ForeignKeys {
		errs = append(errs, validateForeignKey(name, fk, s, columnNames)...)
	}

	// Validate unique constraints
	for _, uc := range t.UniqueConstraints {
		errs = append(errs, validateConstraint(name, "unique constraint", uc.Columns, columnNames)...)
	}

	// Validate indexes
	for _, idx := range t.Indexes {
		errs = append(errs, validateConstraint(name, "index", idx.Columns, columnNames)...)
	}

	return errs
}

func validateColumn(tableName string, c *Column) []error {
	var errs []error

	if c.Name == "" {
		errs = append(errs, fmt.Errorf("table %s: column name cannot be empty", tableName))
	}

	if c.Type == "" {
		errs = append(errs, fmt.Errorf("table %s: column %s: column type cannot be empty", tableName, c.Name))
	}

	// Validate type is a known PostgreSQL type
	if c.Type != "" && !isValidPostgresType(c.Type) {
		errs = append(errs, fmt.Errorf("table %s: column %s: invalid PostgreSQL type: %s", tableName, c.Name, c.Type))
	}

	return errs
}

func validateForeignKey(tableName string, fk *ForeignKey, s *Schema, columnNames map[string]bool) []error {
	var errs []error

	// Check that the column exists in this table
	for _, colName := range fk.Columns {
		if !columnNames[colName] {
			errs = append(errs, fmt.Errorf("table %s: foreign key column %s does not exist", tableName, colName))
		}
	}

	// Check that the referenced table exists
	refTable, exists := s.Tables[fk.ReferencedTable]
	if !exists {
		errs = append(errs, fmt.Errorf("table %s: foreign key references non-existent table %s", tableName, fk.ReferencedTable))
		return errs
	}

	// Check that referenced columns exist
	refColumns := make(map[string]bool)
	for _, col := range refTable.Columns {
		refColumns[col.Name] = true
	}

	for _, refColName := range fk.ReferencedColumns {
		if !refColumns[refColName] {
			errs = append(errs, fmt.Errorf("table %s: foreign key references non-existent column %s in table %s", tableName, refColName, fk.ReferencedTable))
		}
	}

	return errs
}

func validateConstraint(tableName, constraintType string, columns []string, columnNames map[string]bool) []error {
	var errs []error

	for _, colName := range columns {
		if !columnNames[colName] {
			errs = append(errs, fmt.Errorf("table %s: %s references non-existent column %s", tableName, constraintType, colName))
		}
	}

	return errs
}

func validateDependencies(s *Schema) []error {
	var errs []error

	// Build dependency graph
	deps := make(map[string][]string)
	for tableName, table := range s.Tables {
		for _, fk := range table.ForeignKeys {
			// Skip self-references (allowed pattern)
			if fk.ReferencedTable != tableName {
				deps[tableName] = append(deps[tableName], fk.ReferencedTable)
			}
		}
	}

	// Detect cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range deps[node] {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for table := range s.Tables {
		if !visited[table] {
			if hasCycle(table) {
				errs = append(errs, fmt.Errorf("circular dependency detected in foreign key relationships"))
				break
			}
		}
	}

	return errs
}

// isValidPostgresType checks if the type is a valid PostgreSQL type
func isValidPostgresType(typeName string) bool {
	// Extract base type (handle varchar(255), decimal(10,2), etc.)
	baseType := strings.Split(strings.ToLower(typeName), "(")[0]
	baseType = strings.TrimSpace(baseType)

	validTypes := map[string]bool{
		// Integer types
		"smallint": true, "integer": true, "int": true, "bigint": true,
		"serial": true, "bigserial": true, "smallserial": true,

		// Floating-point types
		"real": true, "double precision": true, "numeric": true, "decimal": true,

		// Character types
		"char": true, "varchar": true, "character varying": true, "text": true,

		// Boolean type
		"boolean": true, "bool": true,

		// Date/Time types
		"date": true, "time": true, "timestamp": true, "timestamptz": true,
		"timestamp with time zone": true, "timestamp without time zone": true,
		"interval": true,

		// Binary types
		"bytea": true,

		// UUID type
		"uuid": true,

		// JSON types
		"json": true, "jsonb": true,

		// Array types (basic check)
		"array": true,

		// Network types
		"inet": true, "cidr": true, "macaddr": true,

		// Geometric types
		"point": true, "line": true, "lseg": true, "box": true, "path": true,
		"polygon": true, "circle": true,

		// Money type
		"money": true,

		// XML type
		"xml": true,
	}

	return validTypes[baseType]
}