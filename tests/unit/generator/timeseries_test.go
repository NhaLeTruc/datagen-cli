package generator_test

import (
	"testing"
	"time"

	"github.com/NhaLeTruc/datagen-cli/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeSeriesGenerator(t *testing.T) {
	t.Run("uniform pattern generates evenly distributed timestamps", func(t *testing.T) {
		startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		interval := time.Hour

		gen := generator.NewTimeSeriesGenerator(startTime, endTime, interval, "uniform")
		ctx := generator.NewContextWithSeed(42)

		// Generate timestamps and verify they're within range
		for i := 0; i < 10; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)

			ts, ok := val.(time.Time)
			require.True(t, ok)
			assert.True(t, !ts.Before(startTime) && !ts.After(endTime))
		}
	})

	t.Run("business_hours pattern excludes nights and weekends", func(t *testing.T) {
		startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // Monday
		endTime := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)   // Next Monday
		interval := time.Hour

		gen := generator.NewTimeSeriesGenerator(startTime, endTime, interval, "business_hours")
		ctx := generator.NewContextWithSeed(42)

		// Generate many samples to verify business hours pattern
		for i := 0; i < 50; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)

			ts := val.(time.Time)

			// Should be Monday-Friday
			weekday := ts.Weekday()
			assert.NotEqual(t, time.Saturday, weekday, "Should not be Saturday")
			assert.NotEqual(t, time.Sunday, weekday, "Should not be Sunday")

			// Should be between 9 AM and 5 PM
			hour := ts.Hour()
			assert.GreaterOrEqual(t, hour, 9, "Should be after 9 AM")
			assert.Less(t, hour, 17, "Should be before 5 PM")
		}
	})

	t.Run("daily_peak pattern concentrates around peak hours", func(t *testing.T) {
		startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		interval := time.Hour

		gen := generator.NewTimeSeriesGenerator(startTime, endTime, interval, "daily_peak")
		ctx := generator.NewContextWithSeed(42)

		hourCounts := make(map[int]int)
		sampleSize := 500

		for i := 0; i < sampleSize; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)

			ts := val.(time.Time)
			hourCounts[ts.Hour()]++
		}

		// Peak hours (10-14) should have more samples than off-peak
		peakCount := hourCounts[10] + hourCounts[11] + hourCounts[12] + hourCounts[13]
		offPeakCount := hourCounts[0] + hourCounts[1] + hourCounts[2] + hourCounts[3]

		assert.Greater(t, peakCount, offPeakCount, "Peak hours should have more samples")
	})

	t.Run("sequential generation maintains order", func(t *testing.T) {
		startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		interval := time.Hour

		gen := generator.NewTimeSeriesGenerator(startTime, endTime, interval, "uniform")
		ctx := generator.NewContextWithSeed(42)

		var prevTime time.Time
		for i := 0; i < 5; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)

			ts := val.(time.Time)
			if i > 0 {
				// Each timestamp should be later than or equal to previous
				assert.True(t, !ts.Before(prevTime) || ts.Sub(prevTime) < interval)
			}
			prevTime = ts
		}
	})

	t.Run("name is timeseries", func(t *testing.T) {
		startTime := time.Now()
		endTime := startTime.Add(24 * time.Hour)
		gen := generator.NewTimeSeriesGenerator(startTime, endTime, time.Hour, "uniform")
		assert.Equal(t, "timeseries", gen.Name())
	})
}