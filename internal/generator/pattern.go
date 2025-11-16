package generator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/NhaLeTruc/datagen-cli/internal/schema"
)

// PatternGenerator generates values based on template patterns
// Supports placeholders like: {year}, {month}, {sequence:6}, {random:10}, {uuid}
type PatternGenerator struct {
	config       *schema.PatternConfig
	sequenceMap  map[string]int // Track sequences per pattern
	placeholders *regexp.Regexp
}

// NewPatternGenerator creates a pattern-based generator
func NewPatternGenerator(config *schema.PatternConfig) *PatternGenerator {
	return &PatternGenerator{
		config:       config,
		sequenceMap:  make(map[string]int),
		placeholders: regexp.MustCompile(`\{([^}]+)\}`),
	}
}

func (g *PatternGenerator) Name() string {
	return "pattern"
}

func (g *PatternGenerator) Generate(ctx *Context) (interface{}, error) {
	if g.config.Template == "" {
		return nil, fmt.Errorf("pattern template is empty")
	}

	result := g.config.Template
	now := time.Now()

	// Find all placeholders
	matches := g.placeholders.FindAllStringSubmatch(result, -1)

	for _, match := range matches {
		placeholder := match[0] // Full match with braces: {year}
		spec := match[1]        // Content without braces: year

		// Parse the placeholder spec (might have format: name:param)
		parts := strings.SplitN(spec, ":", 2)
		name := parts[0]
		param := ""
		if len(parts) > 1 {
			param = parts[1]
		}

		// Generate value based on placeholder type
		value, err := g.resolvePlaceholder(ctx, name, param, now)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve placeholder %s: %w", placeholder, err)
		}

		result = strings.Replace(result, placeholder, value, 1)
	}

	return result, nil
}

func (g *PatternGenerator) resolvePlaceholder(ctx *Context, name, param string, now time.Time) (string, error) {
	switch name {
	case "year":
		return strconv.Itoa(now.Year()), nil

	case "month":
		if param == "name" {
			return now.Month().String(), nil
		}
		return fmt.Sprintf("%02d", now.Month()), nil

	case "day":
		return fmt.Sprintf("%02d", now.Day()), nil

	case "timestamp":
		return strconv.FormatInt(now.Unix(), 10), nil

	case "sequence":
		// Increment sequence counter
		g.sequenceMap[g.config.Template]++
		seq := g.sequenceMap[g.config.Template]

		// Apply padding if specified
		if param != "" {
			width, err := strconv.Atoi(param)
			if err != nil {
				return "", fmt.Errorf("invalid sequence width: %s", param)
			}
			format := fmt.Sprintf("%%0%dd", width)
			return fmt.Sprintf(format, seq), nil
		}
		return strconv.Itoa(seq), nil

	case "random":
		// Generate random number with specified digits
		digits := 6 // default
		if param != "" {
			d, err := strconv.Atoi(param)
			if err != nil {
				return "", fmt.Errorf("invalid random digits: %s", param)
			}
			digits = d
		}

		// Generate random number with exact digit count
		min := int(math.Pow10(digits - 1))
		max := int(math.Pow10(digits)) - 1
		num := ctx.Rand.Intn(max-min+1) + min
		return strconv.Itoa(num), nil

	case "uuid":
		return ctx.Faker.UUID().V4(), nil

	case "row":
		// Current row number (1-indexed)
		return strconv.Itoa(ctx.RowNumber + 1), nil

	case "table":
		// Current table name
		return ctx.TableName, nil

	case "hex":
		// Random hex string with specified length
		length := 8 // default
		if param != "" {
			l, err := strconv.Atoi(param)
			if err != nil {
				return "", fmt.Errorf("invalid hex length: %s", param)
			}
			length = l
		}
		return generateHex(ctx, length), nil

	case "alpha":
		// Random alphabetic string
		length := 8 // default
		if param != "" {
			l, err := strconv.Atoi(param)
			if err != nil {
				return "", fmt.Errorf("invalid alpha length: %s", param)
			}
			length = l
		}
		return generateAlpha(ctx, length), nil

	case "alphanumeric":
		// Random alphanumeric string
		length := 8 // default
		if param != "" {
			l, err := strconv.Atoi(param)
			if err != nil {
				return "", fmt.Errorf("invalid alphanumeric length: %s", param)
			}
			length = l
		}
		return generateAlphanumeric(ctx, length), nil

	default:
		// Check if it's a custom variable from config
		if g.config.Variables != nil {
			if val, ok := g.config.Variables[name]; ok {
				return fmt.Sprintf("%v", val), nil
			}
		}
		return "", fmt.Errorf("unknown placeholder: %s", name)
	}
}

// Helper functions for generating random strings
func generateHex(ctx *Context, length int) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = hexChars[ctx.Rand.Intn(len(hexChars))]
	}
	return string(result)
}

func generateAlpha(ctx *Context, length int) string {
	const alphaChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = alphaChars[ctx.Rand.Intn(len(alphaChars))]
	}
	return string(result)
}

func generateAlphanumeric(ctx *Context, length int) string {
	const alphanum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = alphanum[ctx.Rand.Intn(len(alphanum))]
	}
	return string(result)
}

// Need to add this import at the top
import "math"
