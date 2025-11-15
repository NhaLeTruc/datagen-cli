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
		errs = append(errs, fmt.Errorf("table %s: row_count must be greater than 0, got %d\n  → Suggestion: Set 'row_count' to a positive integer (e.g., 100)", name, t.RowCount))
	}

	// Validate columns
	if len(t.Columns) == 0 {
		errs = append(errs, fmt.Errorf("table %s: must have at least one column\n  → Suggestion: Add column definitions in the 'columns' array", name))
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
		errs = append(errs, fmt.Errorf("table %s: column name cannot be empty\n  → Suggestion: Provide a 'name' field for each column", tableName))
	}

	if c.Type == "" {
		columnRef := c.Name
		if columnRef == "" {
			columnRef = "<unnamed>"
		}
		errs = append(errs, fmt.Errorf("table %s: column %s: column type cannot be empty\n  → Suggestion: Add a 'type' field (e.g., 'varchar(255)', 'integer', 'timestamp')", tableName, columnRef))
	}

	// Validate type is a known PostgreSQL type
	if c.Type != "" && !isValidPostgresType(c.Type) {
		suggestion := suggestPostgresType(c.Type)
		errs = append(errs, fmt.Errorf("table %s: column %s: invalid PostgreSQL type '%s'\n  → Suggestion: %s", tableName, c.Name, c.Type, suggestion))
	}

	return errs
}

func validateForeignKey(tableName string, fk *ForeignKey, s *Schema, columnNames map[string]bool) []error {
	var errs []error

	// Check that the column exists in this table
	for _, colName := range fk.Columns {
		if !columnNames[colName] {
			// Build list of available columns
			availableCols := make([]string, 0, len(columnNames))
			for col := range columnNames {
				availableCols = append(availableCols, col)
			}
			errs = append(errs, fmt.Errorf("table %s: foreign key column '%s' does not exist\n  → Suggestion: Add column '%s' to table '%s' or use an existing column: %v",
				tableName, colName, colName, tableName, availableCols))
		}
	}

	// Check that the referenced table exists
	refTable, exists := s.Tables[fk.ReferencedTable]
	if !exists {
		// Build list of available tables
		availableTables := make([]string, 0, len(s.Tables))
		for tblName := range s.Tables {
			availableTables = append(availableTables, tblName)
		}
		errs = append(errs, fmt.Errorf("table %s: foreign key references non-existent table '%s'\n  → Suggestion: Create table '%s' first or reference an existing table: %v",
			tableName, fk.ReferencedTable, fk.ReferencedTable, availableTables))
		return errs
	}

	// Check that referenced columns exist
	refColumns := make(map[string]bool)
	for _, col := range refTable.Columns {
		refColumns[col.Name] = true
	}

	for _, refColName := range fk.ReferencedColumns {
		if !refColumns[refColName] {
			// Build list of available columns in referenced table
			availableRefCols := make([]string, 0, len(refColumns))
			for col := range refColumns {
				availableRefCols = append(availableRefCols, col)
			}
			errs = append(errs, fmt.Errorf("table %s: foreign key references non-existent column '%s' in table '%s'\n  → Suggestion: Add column '%s' to table '%s' or use an existing column: %v",
				tableName, refColName, fk.ReferencedTable, refColName, fk.ReferencedTable, availableRefCols))
		}
	}

	return errs
}

func validateConstraint(tableName, constraintType string, columns []string, columnNames map[string]bool) []error {
	var errs []error

	for _, colName := range columns {
		if !columnNames[colName] {
			// Build list of available columns
			availableCols := make([]string, 0, len(columnNames))
			for col := range columnNames {
				availableCols = append(availableCols, col)
			}
			errs = append(errs, fmt.Errorf("table %s: %s references non-existent column '%s'\n  → Suggestion: Add column '%s' to table '%s' or use an existing column: %v",
				tableName, constraintType, colName, colName, tableName, availableCols))
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

	// Detect cycles using DFS with path tracking
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cyclePath []string

	var hasCycle func(string, []string) bool
	hasCycle = func(node string, path []string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, dep := range deps[node] {
			if !visited[dep] {
				if hasCycle(dep, path) {
					return true
				}
			} else if recStack[dep] {
				// Found a cycle - build the cycle path
				cyclePath = append(path, dep)
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for table := range s.Tables {
		if !visited[table] {
			if hasCycle(table, []string{}) {
				// Format the cycle path
				cycleStr := strings.Join(cyclePath, " → ")
				errs = append(errs, fmt.Errorf("circular dependency detected in foreign key relationships\n  → Cycle: %s\n  → Suggestion: Remove one of the foreign key relationships in this cycle, or make the relationship nullable to allow NULL values during data generation", cycleStr))
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

// suggestPostgresType provides helpful suggestions for invalid PostgreSQL types
func suggestPostgresType(typeName string) string {
	baseType := strings.Split(strings.ToLower(typeName), "(")[0]
	baseType = strings.TrimSpace(baseType)

	// Common mistakes and their suggestions
	suggestions := map[string]string{
		"string":        "Use 'varchar(255)' or 'text' for string data",
		"str":           "Use 'varchar(255)' or 'text' for string data",
		"number":        "Use 'integer' for whole numbers or 'numeric(10,2)' for decimals",
		"float":         "Use 'real' or 'double precision' for floating-point numbers",
		"double":        "Use 'double precision' for 64-bit floating-point numbers",
		"datetime":      "Use 'timestamp' or 'timestamp with time zone' for date/time data",
		"date_time":     "Use 'timestamp' or 'timestamp with time zone' for date/time data",
		"bool":          "Type 'bool' is valid, but 'boolean' is the standard form",
		"tinyint":       "PostgreSQL doesn't have 'tinyint', use 'smallint' instead",
		"mediumint":     "PostgreSQL doesn't have 'mediumint', use 'integer' instead",
		"long":          "Use 'bigint' for 64-bit integers",
		"longtext":      "Use 'text' for large text data",
		"mediumtext":    "Use 'text' for text data",
		"tinytext":      "Use 'varchar(255)' or 'text' for text data",
		"blob":          "Use 'bytea' for binary data",
		"longblob":      "Use 'bytea' for binary data",
		"mediumblob":    "Use 'bytea' for binary data",
		"tinyblob":      "Use 'bytea' for binary data",
		"auto_increment": "Use 'serial', 'bigserial', or 'smallserial' for auto-incrementing integers",
		"enum":          "PostgreSQL doesn't support inline ENUM types in this context, use 'varchar' with CHECK constraint or create a custom ENUM type",
		"set":           "PostgreSQL doesn't have 'SET' type, use 'text[]' (array) instead",
		"year":          "Use 'smallint' or 'integer' for year values",
	}

	if suggestion, exists := suggestions[baseType]; exists {
		return suggestion
	}

	// Check for similar valid types (fuzzy matching)
	validTypes := []string{
		"smallint", "integer", "bigint", "serial", "bigserial", "smallserial",
		"real", "double precision", "numeric", "decimal",
		"varchar", "char", "text",
		"boolean",
		"date", "time", "timestamp", "timestamptz", "interval",
		"bytea", "uuid", "json", "jsonb",
		"inet", "cidr", "macaddr",
		"point", "line", "box", "path", "polygon", "circle",
		"money", "xml",
	}

	// Find similar types using simple string distance
	for _, valid := range validTypes {
		if strings.Contains(baseType, valid) || strings.Contains(valid, baseType) {
			return fmt.Sprintf("Did you mean '%s'? Check the PostgreSQL documentation for valid types", valid)
		}
	}

	return "Check the PostgreSQL documentation for valid data types (https://www.postgresql.org/docs/current/datatype.html)"
}