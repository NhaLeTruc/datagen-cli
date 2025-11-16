package schema

// DistributionConfig represents weighted distribution for column values
type DistributionConfig struct {
	// Type of distribution: "weighted", "normal", "poisson", "zipf"
	Type string `json:"type"`

	// For weighted distribution: map of value -> weight
	// Example: {"completed": 80, "pending": 15, "cancelled": 5}
	Weights map[string]interface{} `json:"weights,omitempty"`

	// For normal distribution
	Mean   *float64 `json:"mean,omitempty"`
	StdDev *float64 `json:"std_dev,omitempty"`

	// For zipf distribution (power-law, e.g., popularity)
	Alpha *float64 `json:"alpha,omitempty"` // Exponent parameter (typically 1.0-2.0)

	// Common parameters
	Min interface{} `json:"min,omitempty"`
	Max interface{} `json:"max,omitempty"`
}

// BusinessRule represents conditional logic for data generation
type BusinessRule struct {
	// Condition to check (evaluated against other columns in same row)
	Condition map[string]interface{} `json:"if"`

	// Action to take if condition is true
	Then map[string]interface{} `json:"then"`

	// Optional else clause
	Else map[string]interface{} `json:"else,omitempty"`
}

// PatternConfig represents custom pattern template configuration
type PatternConfig struct {
	// Template string with placeholders
	// Example: "ACC-{year}-{sequence:6}" -> "ACC-2024-000123"
	Template string `json:"template"`

	// Available variables (computed or custom)
	Variables map[string]interface{} `json:"variables,omitempty"`
}
