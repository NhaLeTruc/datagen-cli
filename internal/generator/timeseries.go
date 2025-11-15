package generator

import (
	"fmt"
	"time"
)

// TimeSeriesGenerator generates time-series data with various patterns
type TimeSeriesGenerator struct {
	startTime time.Time
	endTime   time.Time
	interval  time.Duration
	pattern   string
}

// NewTimeSeriesGenerator creates a time-series generator
func NewTimeSeriesGenerator(startTime, endTime time.Time, interval time.Duration, pattern string) *TimeSeriesGenerator {
	return &TimeSeriesGenerator{
		startTime: startTime,
		endTime:   endTime,
		interval:  interval,
		pattern:   pattern,
	}
}

func (g *TimeSeriesGenerator) Generate(ctx *Context) (interface{}, error) {
	switch g.pattern {
	case "uniform":
		return g.generateUniform(ctx), nil
	case "business_hours":
		return g.generateBusinessHours(ctx), nil
	case "daily_peak":
		return g.generateDailyPeak(ctx), nil
	default:
		return g.generateUniform(ctx), nil
	}
}

func (g *TimeSeriesGenerator) generateUniform(ctx *Context) time.Time {
	// Get current sequence number
	key := fmt.Sprintf("timeseries_%p", g)
	val, exists := ctx.Get(key)
	var seq int64
	if exists {
		seq = val.(int64)
	}

	// Calculate time based on sequence and interval
	duration := time.Duration(seq) * g.interval
	timestamp := g.startTime.Add(duration)

	// If we've exceeded end time, start over with some randomness
	if timestamp.After(g.endTime) {
		totalDuration := g.endTime.Sub(g.startTime)
		randomOffset := time.Duration(ctx.Rand.Int63n(int64(totalDuration)))
		timestamp = g.startTime.Add(randomOffset)
	}

	seq++
	ctx.Set(key, seq)

	return timestamp
}

func (g *TimeSeriesGenerator) generateBusinessHours(ctx *Context) time.Time {
	for {
		ts := g.generateUniform(ctx)

		// Check if it's a weekday
		weekday := ts.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			continue
		}

		// Check if it's during business hours (9 AM - 5 PM)
		hour := ts.Hour()
		if hour >= 9 && hour < 17 {
			return ts
		}
	}
}

func (g *TimeSeriesGenerator) generateDailyPeak(ctx *Context) time.Time {
	// Generate timestamps with bias toward peak hours (10 AM - 2 PM)
	totalDuration := g.endTime.Sub(g.startTime)

	// Use weighted random to favor peak hours
	r := ctx.Rand.Float64()

	var timestamp time.Time
	if r < 0.6 {
		// 60% chance of peak hours (10-14)
		peakStart := g.startTime
		// Adjust to 10 AM on the same day
		peakStart = time.Date(peakStart.Year(), peakStart.Month(), peakStart.Day(),
			10, 0, 0, 0, peakStart.Location())

		peakDuration := 4 * time.Hour // 10 AM to 2 PM
		offset := time.Duration(ctx.Rand.Int63n(int64(peakDuration)))
		timestamp = peakStart.Add(offset)
	} else {
		// 40% chance of off-peak hours
		offset := time.Duration(ctx.Rand.Int63n(int64(totalDuration)))
		timestamp = g.startTime.Add(offset)

		// Exclude peak hours
		hour := timestamp.Hour()
		if hour >= 10 && hour < 14 {
			// Shift to off-peak
			timestamp = timestamp.Add(6 * time.Hour)
		}
	}

	// Ensure within bounds
	if timestamp.After(g.endTime) {
		timestamp = g.endTime
	}
	if timestamp.Before(g.startTime) {
		timestamp = g.startTime
	}

	return timestamp
}

func (g *TimeSeriesGenerator) Name() string {
	return "timeseries"
}