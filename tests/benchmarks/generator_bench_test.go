package benchmarks

import (
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/generator"
)

// Benchmark basic type generators
func BenchmarkBasicGenerators(b *testing.B) {
	ctx := generator.NewGenerationContext(12345)

	b.Run("IntegerGenerator", func(b *testing.B) {
		gen := &generator.IntegerGenerator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, nil)
		}
	})

	b.Run("VarcharGenerator", func(b *testing.B) {
		gen := &generator.VarcharGenerator{}
		config := map[string]interface{}{"max_length": 255}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, config)
		}
	})

	b.Run("TimestampGenerator", func(b *testing.B) {
		gen := &generator.TimestampGenerator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, nil)
		}
	})

	b.Run("BooleanGenerator", func(b *testing.B) {
		gen := &generator.BooleanGenerator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, nil)
		}
	})

	b.Run("UUIDGenerator", func(b *testing.B) {
		gen := &generator.UUIDGenerator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, nil)
		}
	})
}

// Benchmark semantic generators
func BenchmarkSemanticGenerators(b *testing.B) {
	ctx := generator.NewGenerationContext(12345)

	b.Run("EmailGenerator", func(b *testing.B) {
		gen := &generator.EmailGenerator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, nil)
		}
	})

	b.Run("PhoneGenerator", func(b *testing.B) {
		gen := &generator.PhoneGenerator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, nil)
		}
	})

	b.Run("NameGenerator", func(b *testing.B) {
		gen := &generator.FullNameGenerator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, nil)
		}
	})

	b.Run("AddressGenerator", func(b *testing.B) {
		gen := &generator.AddressGenerator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, nil)
		}
	})
}

// Benchmark custom pattern generators
func BenchmarkCustomGenerators(b *testing.B) {
	ctx := generator.NewGenerationContext(12345)

	b.Run("WeightedEnumGenerator", func(b *testing.B) {
		gen := &generator.WeightedEnumGenerator{}
		config := map[string]interface{}{
			"values":  []interface{}{"active", "inactive", "suspended"},
			"weights": []interface{}{80, 15, 5},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, config)
		}
	})

	b.Run("PatternGenerator", func(b *testing.B) {
		gen := &generator.PatternGenerator{}
		config := map[string]interface{}{
			"pattern": "PRD-[0-9]{6}",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, config)
		}
	})

	b.Run("TemplateGenerator", func(b *testing.B) {
		gen := &generator.TemplateGenerator{}
		config := map[string]interface{}{
			"template": "ORD-{{year}}-{{seq:8}}",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, config)
		}
	})

	b.Run("IntegerRangeGenerator", func(b *testing.B) {
		gen := &generator.IntegerRangeGenerator{}
		config := map[string]interface{}{
			"min": 18,
			"max": 100,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, config)
		}
	})
}

// Benchmark time-series generators
func BenchmarkTimeseriesGenerators(b *testing.B) {
	ctx := generator.NewGenerationContext(12345)

	b.Run("UniformTimeseries", func(b *testing.B) {
		gen := &generator.TimeseriesGenerator{}
		config := map[string]interface{}{
			"start":   "2024-01-01T00:00:00Z",
			"end":     "2024-12-31T23:59:59Z",
			"pattern": "uniform",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, config)
		}
	})

	b.Run("BusinessHoursTimeseries", func(b *testing.B) {
		gen := &generator.TimeseriesGenerator{}
		config := map[string]interface{}{
			"start":   "2024-01-01T00:00:00Z",
			"end":     "2024-12-31T23:59:59Z",
			"pattern": "business_hours",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, config)
		}
	})

	b.Run("DailyPeakTimeseries", func(b *testing.B) {
		gen := &generator.TimeseriesGenerator{}
		config := map[string]interface{}{
			"start":      "2024-01-01T00:00:00Z",
			"end":        "2024-12-31T23:59:59Z",
			"pattern":    "daily_peak",
			"peak_hours": []interface{}{12, 13, 18, 19},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = gen.Generate(ctx, config)
		}
	})
}

// Benchmark generator registry operations
func BenchmarkGeneratorRegistry(b *testing.B) {
	b.Run("RegisterGenerator", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gen := &generator.IntegerGenerator{}
			generator.Register(gen)
		}
	})

	b.Run("GetGenerator", func(b *testing.B) {
		// Register a generator first
		gen := &generator.IntegerGenerator{}
		generator.Register(gen)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = generator.Get("integer")
		}
	})
}

// Benchmark bulk generation (realistic workload)
func BenchmarkBulkGeneration(b *testing.B) {
	ctx := generator.NewGenerationContext(12345)

	b.Run("Generate1000Emails", func(b *testing.B) {
		gen := &generator.EmailGenerator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				_, _ = gen.Generate(ctx, nil)
			}
		}
	})

	b.Run("Generate1000Timestamps", func(b *testing.B) {
		gen := &generator.TimestampGenerator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				_, _ = gen.Generate(ctx, nil)
			}
		}
	})

	b.Run("Generate1000WeightedEnums", func(b *testing.B) {
		gen := &generator.WeightedEnumGenerator{}
		config := map[string]interface{}{
			"values":  []interface{}{"active", "inactive", "suspended"},
			"weights": []interface{}{80, 15, 5},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000; j++ {
				_, _ = gen.Generate(ctx, config)
			}
		}
	})
}

// Benchmark context creation
func BenchmarkGenerationContext(b *testing.B) {
	b.Run("CreateContext", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = generator.NewGenerationContext(12345)
		}
	})

	b.Run("CreateContextWithSeed", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			seed := int64(i)
			_ = generator.NewGenerationContext(seed)
		}
	})
}
