package benchmarks

import (
	"bytes"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/NhaLeTruc/datagen-cli/internal/schema"
)

// Benchmark schema parsing
func BenchmarkSchemaParsing(b *testing.B) {
	simpleSchema := `{
		"version": "1.0",
		"database": {"name": "test_db"},
		"tables": {
			"users": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "email", "type": "varchar(255)", "nullable": false}
				],
				"primary_key": ["id"],
				"row_count": 100
			}
		}
	}`

	b.Run("ParseSimpleSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = schema.Parse([]byte(simpleSchema))
		}
	})

	complexSchema := `{
		"version": "1.0",
		"database": {"name": "test_db"},
		"tables": {
			"users": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "email", "type": "varchar(255)", "nullable": false},
					{"name": "first_name", "type": "varchar(100)", "nullable": false},
					{"name": "last_name", "type": "varchar(100)", "nullable": false},
					{"name": "age", "type": "integer", "nullable": true},
					{"name": "created_at", "type": "timestamptz", "nullable": false}
				],
				"primary_key": ["id"],
				"row_count": 1000
			},
			"posts": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "user_id", "type": "integer", "nullable": false},
					{"name": "title", "type": "varchar(255)", "nullable": false},
					{"name": "content", "type": "text", "nullable": false},
					{"name": "created_at", "type": "timestamptz", "nullable": false}
				],
				"primary_key": ["id"],
				"foreign_keys": [
					{
						"columns": ["user_id"],
						"referenced_table": "users",
						"referenced_columns": ["id"]
					}
				],
				"row_count": 5000
			}
		}
	}`

	b.Run("ParseComplexSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = schema.Parse([]byte(complexSchema))
		}
	})
}

// Benchmark schema validation
func BenchmarkSchemaValidation(b *testing.B) {
	schemaJSON := `{
		"version": "1.0",
		"database": {"name": "test_db"},
		"tables": {
			"users": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "email", "type": "varchar(255)", "nullable": false}
				],
				"primary_key": ["id"],
				"row_count": 100
			}
		}
	}`

	sch, _ := schema.Parse([]byte(schemaJSON))

	b.Run("ValidateSimpleSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.Validate(sch)
		}
	})
}

// Benchmark dependency resolution
func BenchmarkDependencyResolution(b *testing.B) {
	schemaJSON := `{
		"version": "1.0",
		"database": {"name": "test_db"},
		"tables": {
			"users": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false}
				],
				"primary_key": ["id"],
				"row_count": 100
			},
			"posts": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "user_id", "type": "integer", "nullable": false}
				],
				"primary_key": ["id"],
				"foreign_keys": [
					{
						"columns": ["user_id"],
						"referenced_table": "users",
						"referenced_columns": ["id"]
					}
				],
				"row_count": 500
			},
			"comments": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "post_id", "type": "integer", "nullable": false},
					{"name": "user_id", "type": "integer", "nullable": false}
				],
				"primary_key": ["id"],
				"foreign_keys": [
					{
						"columns": ["post_id"],
						"referenced_table": "posts",
						"referenced_columns": ["id"]
					},
					{
						"columns": ["user_id"],
						"referenced_table": "users",
						"referenced_columns": ["id"]
					}
				],
				"row_count": 2000
			}
		}
	}`

	sch, _ := schema.Parse([]byte(schemaJSON))

	b.Run("ResolveDependencies", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pipeline.ResolveDependencies(sch)
		}
	})
}

// Benchmark full pipeline (end-to-end)
func BenchmarkFullPipeline(b *testing.B) {
	schemaJSON := `{
		"version": "1.0",
		"database": {"name": "test_db"},
		"tables": {
			"users": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "email", "type": "varchar(255)", "nullable": false},
					{"name": "created_at", "type": "timestamptz", "nullable": false}
				],
				"primary_key": ["id"],
				"row_count": 100
			}
		}
	}`

	b.Run("GenerateSmallDataset_SQL", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sch, _ := schema.Parse([]byte(schemaJSON))
			output := new(bytes.Buffer)
			_ = pipeline.Generate(sch, output, "sql", 12345)
		}
	})

	b.Run("GenerateSmallDataset_COPY", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sch, _ := schema.Parse([]byte(schemaJSON))
			output := new(bytes.Buffer)
			_ = pipeline.Generate(sch, output, "copy", 12345)
		}
	})
}

// Benchmark throughput (rows/second)
func BenchmarkThroughput(b *testing.B) {
	// Small table (100 rows)
	smallSchema := `{
		"version": "1.0",
		"database": {"name": "test_db"},
		"tables": {
			"users": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "email", "type": "varchar(255)", "nullable": false}
				],
				"primary_key": ["id"],
				"row_count": 100
			}
		}
	}`

	// Medium table (1000 rows)
	mediumSchema := `{
		"version": "1.0",
		"database": {"name": "test_db"},
		"tables": {
			"users": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "email", "type": "varchar(255)", "nullable": false},
					{"name": "first_name", "type": "varchar(100)", "nullable": false},
					{"name": "last_name", "type": "varchar(100)", "nullable": false}
				],
				"primary_key": ["id"],
				"row_count": 1000
			}
		}
	}`

	// Large table (10000 rows)
	largeSchema := `{
		"version": "1.0",
		"database": {"name": "test_db"},
		"tables": {
			"users": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "email", "type": "varchar(255)", "nullable": false},
					{"name": "first_name", "type": "varchar(100)", "nullable": false},
					{"name": "last_name", "type": "varchar(100)", "nullable": false},
					{"name": "age", "type": "integer", "nullable": true},
					{"name": "created_at", "type": "timestamptz", "nullable": false}
				],
				"primary_key": ["id"],
				"row_count": 10000
			}
		}
	}`

	b.Run("Throughput_100rows", func(b *testing.B) {
		sch, _ := schema.Parse([]byte(smallSchema))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			output := new(bytes.Buffer)
			_ = pipeline.Generate(sch, output, "sql", 12345)
		}
		// Report rows/op metric
		b.ReportMetric(100, "rows/op")
	})

	b.Run("Throughput_1000rows", func(b *testing.B) {
		sch, _ := schema.Parse([]byte(mediumSchema))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			output := new(bytes.Buffer)
			_ = pipeline.Generate(sch, output, "sql", 12345)
		}
		b.ReportMetric(1000, "rows/op")
	})

	b.Run("Throughput_10000rows", func(b *testing.B) {
		sch, _ := schema.Parse([]byte(largeSchema))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			output := new(bytes.Buffer)
			_ = pipeline.Generate(sch, output, "sql", 12345)
		}
		b.ReportMetric(10000, "rows/op")
	})
}

// Benchmark memory allocations
func BenchmarkMemoryAllocations(b *testing.B) {
	schemaJSON := `{
		"version": "1.0",
		"database": {"name": "test_db"},
		"tables": {
			"users": {
				"columns": [
					{"name": "id", "type": "serial", "nullable": false},
					{"name": "email", "type": "varchar(255)", "nullable": false}
				],
				"primary_key": ["id"],
				"row_count": 1000
			}
		}
	}`

	b.Run("AllocationsPerRow", func(b *testing.B) {
		sch, _ := schema.Parse([]byte(schemaJSON))
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			output := new(bytes.Buffer)
			_ = pipeline.Generate(sch, output, "sql", 12345)
		}
	})
}
