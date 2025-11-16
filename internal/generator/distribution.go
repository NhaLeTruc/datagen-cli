package generator

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/NhaLeTruc/datagen-cli/internal/schema"
)

// DistributionGenerator generates values based on weighted distributions
type DistributionGenerator struct {
	config *schema.DistributionConfig
}

// NewDistributionGenerator creates a distribution-based generator
func NewDistributionGenerator(config *schema.DistributionConfig) *DistributionGenerator {
	return &DistributionGenerator{config: config}
}

func (g *DistributionGenerator) Name() string {
	return "distribution"
}

func (g *DistributionGenerator) Generate(ctx *Context) (interface{}, error) {
	switch g.config.Type {
	case "weighted":
		return g.generateWeighted(ctx)
	case "normal":
		return g.generateNormal(ctx)
	case "poisson":
		return g.generatePoisson(ctx)
	case "zipf":
		return g.generateZipf(ctx)
	default:
		return nil, fmt.Errorf("unknown distribution type: %s", g.config.Type)
	}
}

// generateWeighted generates a value based on weighted probabilities
func (g *DistributionGenerator) generateWeighted(ctx *Context) (interface{}, error) {
	if g.config.Weights == nil || len(g.config.Weights) == 0 {
		return nil, fmt.Errorf("weights not specified for weighted distribution")
	}

	// Calculate total weight
	totalWeight := 0.0
	weights := make([]weightedValue, 0, len(g.config.Weights))

	for value, weight := range g.config.Weights {
		w := toFloat64(weight)
		totalWeight += w
		weights = append(weights, weightedValue{
			value:  value,
			weight: w,
		})
	}

	// Sort by weight for deterministic behavior
	sort.Slice(weights, func(i, j int) bool {
		return weights[i].value < weights[j].value
	})

	// Select value based on weighted random
	r := ctx.Rand.Float64() * totalWeight
	cumulative := 0.0

	for _, wv := range weights {
		cumulative += wv.weight
		if r <= cumulative {
			return wv.value, nil
		}
	}

	// Fallback to last value (should rarely happen due to floating point)
	return weights[len(weights)-1].value, nil
}

// generateNormal generates a value from normal (Gaussian) distribution
func (g *DistributionGenerator) generateNormal(ctx *Context) (interface{}, error) {
	if g.config.Mean == nil || g.config.StdDev == nil {
		return nil, fmt.Errorf("mean and std_dev required for normal distribution")
	}

	// Box-Muller transform for normal distribution
	value := ctx.Rand.NormFloat64()*(*g.config.StdDev) + (*g.config.Mean)

	// Apply bounds if specified
	if g.config.Min != nil {
		minVal := toFloat64(g.config.Min)
		if value < minVal {
			value = minVal
		}
	}
	if g.config.Max != nil {
		maxVal := toFloat64(g.config.Max)
		if value > maxVal {
			value = maxVal
		}
	}

	return value, nil
}

// generatePoisson generates a value from Poisson distribution
// Useful for modeling counts, arrivals, events per time period
func (g *DistributionGenerator) generatePoisson(ctx *Context) (interface{}, error) {
	if g.config.Mean == nil {
		return nil, fmt.Errorf("mean (lambda) required for Poisson distribution")
	}

	lambda := *g.config.Mean
	if lambda <= 0 {
		return nil, fmt.Errorf("lambda must be positive for Poisson distribution")
	}

	// Knuth's algorithm for Poisson
	L := math.Exp(-lambda)
	k := 0
	p := 1.0

	for p > L {
		k++
		p *= ctx.Rand.Float64()
	}

	return k - 1, nil
}

// generateZipf generates a value from Zipf distribution (power-law)
// Useful for modeling popularity, word frequency, city sizes
func (g *DistributionGenerator) generateZipf(ctx *Context) (interface{}, error) {
	if g.config.Alpha == nil {
		return nil, fmt.Errorf("alpha required for Zipf distribution")
	}

	// Default range: 1 to 1000
	min := 1
	max := 1000

	if g.config.Min != nil {
		min = int(toFloat64(g.config.Min))
	}
	if g.config.Max != nil {
		max = int(toFloat64(g.config.Max))
	}

	// Use Go's built-in Zipf generator
	zipf := rand.NewZipf(ctx.Rand, *g.config.Alpha, 1.0, uint64(max-min))
	return int(zipf.Uint64()) + min, nil
}

// weightedValue represents a value with its weight
type weightedValue struct {
	value  string
	weight float64
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return 0
	}
}
