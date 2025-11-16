package pipeline_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerPool(t *testing.T) {
	tests := []struct {
		name        string
		workers     int
		tasks       int
		wantResults int
	}{
		{
			name:        "single worker processes all tasks",
			workers:     1,
			tasks:       10,
			wantResults: 10,
		},
		{
			name:        "multiple workers process tasks concurrently",
			workers:     4,
			tasks:       20,
			wantResults: 20,
		},
		{
			name:        "more workers than tasks",
			workers:     10,
			tasks:       5,
			wantResults: 5,
		},
		{
			name:        "no tasks",
			workers:     4,
			tasks:       0,
			wantResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := pipeline.NewWorkerPool(tt.workers)
			require.NotNil(t, pool)

			ctx := context.Background()
			var mu sync.Mutex
			results := make([]int, 0, tt.tasks)

			// Start the pool
			pool.Start(ctx)

			// Submit tasks
			for i := 0; i < tt.tasks; i++ {
				taskNum := i
				pool.Submit(func() error {
					// Simulate some work
					time.Sleep(10 * time.Millisecond)
					mu.Lock()
					results = append(results, taskNum)
					mu.Unlock()
					return nil
				})
			}

			// Wait for completion
			err := pool.Wait()
			require.NoError(t, err)

			// Verify all tasks were processed
			assert.Len(t, results, tt.wantResults)
		})
	}
}

func TestWorkerPoolConcurrency(t *testing.T) {
	workers := 4
	tasks := 100
	pool := pipeline.NewWorkerPool(workers)

	ctx := context.Background()
	pool.Start(ctx)

	var mu sync.Mutex
	concurrent := 0
	maxConcurrent := 0
	completed := 0

	for i := 0; i < tasks; i++ {
		pool.Submit(func() error {
			mu.Lock()
			concurrent++
			if concurrent > maxConcurrent {
				maxConcurrent = concurrent
			}
			mu.Unlock()

			// Simulate work
			time.Sleep(20 * time.Millisecond)

			mu.Lock()
			concurrent--
			completed++
			mu.Unlock()

			return nil
		})
	}

	err := pool.Wait()
	require.NoError(t, err)

	// Verify all tasks completed
	assert.Equal(t, tasks, completed)

	// Verify concurrency was limited to worker count
	assert.LessOrEqual(t, maxConcurrent, workers,
		"max concurrent tasks should not exceed worker count")
}

func TestWorkerPoolBackpressure(t *testing.T) {
	// Test that submitting many tasks doesn't cause unbounded memory growth
	workers := 2
	tasks := 1000
	pool := pipeline.NewWorkerPool(workers)

	ctx := context.Background()
	pool.Start(ctx)

	var completed int
	var mu sync.Mutex

	// Submit many tasks quickly
	for i := 0; i < tasks; i++ {
		pool.Submit(func() error {
			time.Sleep(5 * time.Millisecond)
			mu.Lock()
			completed++
			mu.Unlock()
			return nil
		})
	}

	err := pool.Wait()
	require.NoError(t, err)

	assert.Equal(t, tasks, completed)
}

func TestWorkerPoolErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		tasks         int
		errorOnTask   int
		wantErr       bool
		continueOnErr bool
	}{
		{
			name:          "error stops processing by default",
			tasks:         10,
			errorOnTask:   5,
			wantErr:       true,
			continueOnErr: false,
		},
		{
			name:          "continue on error when configured",
			tasks:         10,
			errorOnTask:   5,
			wantErr:       true,
			continueOnErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := pipeline.NewWorkerPool(4)
			if tt.continueOnErr {
				pool.SetContinueOnError(true)
			}

			ctx := context.Background()
			pool.Start(ctx)

			var mu sync.Mutex
			completed := 0

			for i := 0; i < tt.tasks; i++ {
				taskNum := i
				pool.Submit(func() error {
					if taskNum == tt.errorOnTask {
						return assert.AnError
					}
					mu.Lock()
					completed++
					mu.Unlock()
					return nil
				})
			}

			err := pool.Wait()

			if tt.wantErr {
				assert.Error(t, err)
			}

			if tt.continueOnErr {
				// Should complete all non-error tasks
				assert.Equal(t, tt.tasks-1, completed)
			}
		})
	}
}

func TestWorkerPoolContext(t *testing.T) {
	t.Run("cancellation stops workers", func(t *testing.T) {
		pool := pipeline.NewWorkerPool(4)

		ctx, cancel := context.WithCancel(context.Background())
		pool.Start(ctx)

		var mu sync.Mutex
		started := 0
		completed := 0

		// Submit long-running tasks
		for i := 0; i < 20; i++ {
			pool.Submit(func() error {
				mu.Lock()
				started++
				mu.Unlock()

				// Long task
				time.Sleep(100 * time.Millisecond)

				mu.Lock()
				completed++
				mu.Unlock()
				return nil
			})
		}

		// Cancel after a short time
		time.Sleep(50 * time.Millisecond)
		cancel()

		err := pool.Wait()
		assert.Error(t, err) // Should return context error

		// Some tasks may have started, but not all should complete
		assert.Greater(t, started, 0)
		assert.Less(t, completed, 20)
	})

	t.Run("timeout cancels workers", func(t *testing.T) {
		pool := pipeline.NewWorkerPool(2)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		pool.Start(ctx)

		var completed int
		var mu sync.Mutex

		// Submit tasks that would take too long
		for i := 0; i < 10; i++ {
			pool.Submit(func() error {
				time.Sleep(200 * time.Millisecond)
				mu.Lock()
				completed++
				mu.Unlock()
				return nil
			})
		}

		err := pool.Wait()
		assert.Error(t, err)
		assert.Less(t, completed, 10)
	})
}

func TestWorkerPoolReuse(t *testing.T) {
	t.Run("cannot reuse pool after wait", func(t *testing.T) {
		pool := pipeline.NewWorkerPool(4)

		ctx := context.Background()
		pool.Start(ctx)

		pool.Submit(func() error {
			return nil
		})

		err := pool.Wait()
		require.NoError(t, err)

		// Attempting to submit after Wait should panic or error
		assert.Panics(t, func() {
			pool.Submit(func() error {
				return nil
			})
		})
	})
}

func TestWorkerPoolStats(t *testing.T) {
	t.Run("reports accurate statistics", func(t *testing.T) {
		pool := pipeline.NewWorkerPool(4)

		ctx := context.Background()
		pool.Start(ctx)

		tasks := 20
		for i := 0; i < tasks; i++ {
			pool.Submit(func() error {
				time.Sleep(10 * time.Millisecond)
				return nil
			})
		}

		err := pool.Wait()
		require.NoError(t, err)

		stats := pool.Stats()
		assert.Equal(t, tasks, stats.TotalTasks)
		assert.Equal(t, tasks, stats.CompletedTasks)
		assert.Equal(t, 0, stats.FailedTasks)
		assert.Greater(t, stats.Duration, time.Duration(0))
	})
}
