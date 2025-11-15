package schema

import (
	"encoding/json"
	"fmt"
	"io"
)

// Parse reads a JSON schema from the given reader and returns a Schema struct.
// It validates that required fields (version, database) are present.
func Parse(r io.Reader) (*Schema, error) {
	var s Schema
	decoder := json.NewDecoder(r)

	if err := decoder.Decode(&s); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	if err := validateRequired(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// validateRequired ensures required fields are present
func validateRequired(s *Schema) error {
	if s.Version == "" {
		return fmt.Errorf("required field 'version' is missing or empty")
	}
	if s.Database.Name == "" {
		return fmt.Errorf("required field 'database.name' is missing or empty")
	}
	return nil
}