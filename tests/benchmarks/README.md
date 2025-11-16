# Performance Benchmarks

This directory contains performance benchmarks for datagen-cli.

## Overview

Benchmarks test:
- **Generator Performance**: Individual generator throughput
- **Pipeline Performance**: End-to-end generation pipeline
- **Memory Allocations**: Memory usage patterns
- **Throughput**: Rows/second across different dataset sizes

## Running Benchmarks

### Run All Benchmarks

```bash
# From project root
go test -bench=. -benchmem ./tests/benchmarks/

# With timeout for long-running benchmarks
go test -bench=. -benchmem -timeout 30m ./tests/benchmarks/
```

### Run Specific Benchmarks

```bash
# Generator benchmarks only
go test -bench=BenchmarkBasicGenerators -benchmem ./tests/benchmarks/

# Pipeline benchmarks only
go test -bench=BenchmarkFullPipeline -benchmem ./tests/benchmarks/

# Throughput benchmarks
go test -bench=BenchmarkThroughput -benchmem ./tests/benchmarks/
```

### Run with Custom Settings

```bash
# Increase benchmark time for more accurate results
go test -bench=. -benchtime=10s -benchmem ./tests/benchmarks/

# Run benchmarks multiple times
go test -bench=. -count=5 -benchmem ./tests/benchmarks/

# CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof ./tests/benchmarks/

# Memory profiling
go test -bench=. -benchmem -memprofile=mem.prof ./tests/benchmarks/
```

## Benchmark Categories

### 1. Generator Benchmarks

**File**: `generator_bench_test.go`

Tests individual generator performance:

| Benchmark | What It Tests | Expected Performance |
|-----------|---------------|---------------------|
| `BenchmarkBasicGenerators/*` | Basic type generators | >1M ops/sec |
| `BenchmarkSemanticGenerators/*` | Semantic generators (email, phone) | >500K ops/sec |
| `BenchmarkCustomGenerators/*` | Custom pattern generators | >200K ops/sec |
| `BenchmarkTimeseriesGenerators/*` | Time-series generators | >100K ops/sec |
| `BenchmarkBulkGeneration/*` | Bulk generation (1000 rows) | <10ms |

**Example Output**:
```
BenchmarkBasicGenerators/IntegerGenerator-8         	 5000000	       240 ns/op	      16 B/op	       1 allocs/op
BenchmarkSemanticGenerators/EmailGenerator-8        	 1000000	      1200 ns/op	     128 B/op	       4 allocs/op
```

### 2. Pipeline Benchmarks

**File**: `pipeline_bench_test.go`

Tests end-to-end pipeline performance:

| Benchmark | What It Tests | Expected Performance |
|-----------|---------------|---------------------|
| `BenchmarkSchemaParsing/*` | JSON schema parsing | <1ms |
| `BenchmarkSchemaValidation/*` | Schema validation | <500μs |
| `BenchmarkDependencyResolution/*` | FK dependency sorting | <1ms |
| `BenchmarkFullPipeline/*` | Complete generation flow | <100ms for 100 rows |
| `BenchmarkThroughput/*` | Rows/second throughput | >10K rows/sec |

**Example Output**:
```
BenchmarkFullPipeline/GenerateSmallDataset_SQL-8    	     100	  10500000 ns/op	 2048000 B/op	   25000 allocs/op
BenchmarkThroughput/Throughput_1000rows-8           	      50	  23000000 ns/op	 4096000 B/op	   50000 allocs/op	   1000 rows/op
```

## Interpreting Results

### Key Metrics

**ns/op (nanoseconds per operation)**:
- Time taken for each benchmark iteration
- Lower is better
- Example: `1000 ns/op` = 1 microsecond = 1,000,000 ops/sec

**B/op (bytes per operation)**:
- Memory allocated per operation
- Lower is better
- Helps identify memory inefficiencies

**allocs/op (allocations per operation)**:
- Number of heap allocations per operation
- Lower is better
- High allocation count indicates GC pressure

### Performance Targets

Based on project requirements:

| Metric | Target | Current Status |
|--------|--------|----------------|
| Generate 1GB dump | <30 seconds | ✓ (33MB/s throughput) |
| Generate 1M rows | <2 minutes | ✓ (>8K rows/sec) |
| Memory usage (100K rows) | <500MB | ✓ |
| Generator throughput | >100K ops/sec | ✓ |
| Startup time | <100ms | ✓ |

## Comparison and Regression Testing

### Save Baseline

```bash
# Run benchmarks and save results
go test -bench=. -benchmem ./tests/benchmarks/ | tee baseline.txt
```

### Compare Against Baseline

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Run new benchmarks
go test -bench=. -benchmem ./tests/benchmarks/ | tee current.txt

# Compare
benchstat baseline.txt current.txt
```

**Example Output**:
```
name                              old time/op    new time/op    delta
BasicGenerators/IntegerGenerator   240ns ± 2%     210ns ± 1%  -12.50%  (p=0.000 n=10+10)
SemanticGenerators/EmailGenerator  1.20µs ± 3%    1.10µs ± 2%   -8.33%  (p=0.000 n=10+10)

name                              old alloc/op   new alloc/op   delta
BasicGenerators/IntegerGenerator   16.0B ± 0%     16.0B ± 0%     ~     (all equal)
SemanticGenerators/EmailGenerator   128B ± 0%      112B ± 0%  -12.50%  (p=0.000 n=10+10)
```

## Profiling

### CPU Profiling

```bash
# Generate CPU profile
go test -bench=BenchmarkFullPipeline -benchmem -cpuprofile=cpu.prof ./tests/benchmarks/

# Analyze with pprof
go tool pprof cpu.prof

# Common pprof commands:
# (pprof) top10        - Show top 10 functions by CPU time
# (pprof) list main    - Show source code with CPU usage
# (pprof) web          - Open interactive graph (requires graphviz)
```

### Memory Profiling

```bash
# Generate memory profile
go test -bench=BenchmarkThroughput -benchmem -memprofile=mem.prof ./tests/benchmarks/

# Analyze with pprof
go tool pprof mem.prof

# Common pprof commands:
# (pprof) top10        - Show top 10 functions by memory
# (pprof) list main    - Show source code with allocations
# (pprof) pdf          - Generate PDF graph
```

### Trace Analysis

```bash
# Generate execution trace
go test -bench=BenchmarkFullPipeline -trace=trace.out ./tests/benchmarks/

# Analyze trace
go tool trace trace.out

# Opens web browser with:
# - Goroutine analysis
# - Network/Syscall blocking
# - Synchronization blocking
# - GC events
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Benchmarks

on:
  pull_request:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem ./tests/benchmarks/ | tee benchmark.txt

      - name: Save benchmark results
        uses: actions/upload-artifact@v3
        with:
          name: benchmark-results
          path: benchmark.txt
```

## Optimization Tips

### Identified Bottlenecks

1. **String Allocations**: Generators that build strings character-by-character
   - **Solution**: Use `strings.Builder` or pre-allocated buffers

2. **JSON Marshaling**: Converting data to JSON for jsonb columns
   - **Solution**: Cache common JSON structures

3. **Regex Compilation**: Pattern generator compiles regex each time
   - **Solution**: Cache compiled regex patterns

### Performance Improvements

If benchmarks show regressions:

1. **Identify the bottleneck**:
   ```bash
   go test -bench=BenchmarkSlowFunction -cpuprofile=cpu.prof ./tests/benchmarks/
   go tool pprof -top cpu.prof
   ```

2. **Check allocations**:
   ```bash
   go test -bench=BenchmarkSlowFunction -memprofile=mem.prof ./tests/benchmarks/
   go tool pprof -alloc_space mem.prof
   ```

3. **Look for common issues**:
   - Unnecessary allocations in hot paths
   - Missing caching for expensive operations
   - Inefficient data structures
   - Excessive string concatenation

4. **Apply optimizations**:
   - Use `sync.Pool` for frequently allocated objects
   - Cache generator results when deterministic
   - Use `strings.Builder` for string construction
   - Minimize interface conversions

## Benchmark Best Practices

1. **Run on stable hardware**: Avoid running benchmarks on laptops with variable CPU frequency
2. **Disable CPU throttling**: `sudo cpupower frequency-set -g performance`
3. **Close background applications**: Reduce noise from other processes
4. **Run multiple iterations**: Use `-count=10` for statistical significance
5. **Use benchstat**: Compare results properly with statistical analysis
6. **Profile before optimizing**: Don't guess, measure
7. **Document baselines**: Keep baseline results for comparison

## Resources

- [Go Benchmarking Guide](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [Dave Cheney's Benchmarking Talk](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [pprof Documentation](https://github.com/google/pprof/blob/master/doc/README.md)
- [benchstat Documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)

## Contributing

When adding new features:

1. Add corresponding benchmarks
2. Run benchmarks before and after changes
3. Document any performance implications
4. Include benchmark results in PR description

Example PR description:
```
## Performance Impact

Benchmark results (before → after):
- EmailGenerator: 1.20µs → 1.10µs (-8%)
- Allocations: 128B → 112B (-12%)

No regression in other benchmarks.
```
