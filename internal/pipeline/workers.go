package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task represents a unit of work to be executed by the worker pool
type Task func() error

// WorkerPool manages a pool of goroutines that execute tasks concurrently
type WorkerPool struct {
	workers         int
	tasks           chan Task
	errors          chan error
	wg              sync.WaitGroup
	mu              sync.Mutex
	ctx             context.Context
	cancel          context.CancelFunc
	started         bool
	closed          bool
	continueOnError bool
	stats           PoolStats
	startTime       time.Time
}

// PoolStats contains statistics about the worker pool execution
type PoolStats struct {
	TotalTasks     int
	CompletedTasks int
	FailedTasks    int
	Duration       time.Duration
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int) *WorkerPool {
	if workers < 1 {
		workers = 1
	}

	return &WorkerPool{
		workers: workers,
		tasks:   make(chan Task, workers*2), // Buffered to prevent blocking
		errors:  make(chan error, 1),        // Buffered to capture first error
		stats:   PoolStats{},
	}
}

// SetContinueOnError configures whether the pool should continue processing
// tasks after encountering an error
func (wp *WorkerPool) SetContinueOnError(continueOnErr bool) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.continueOnError = continueOnErr
}

// Start initializes the worker pool and starts the worker goroutines
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.started {
		panic("worker pool already started")
	}

	wp.ctx, wp.cancel = context.WithCancel(ctx)
	wp.started = true
	wp.startTime = time.Now()

	// Start worker goroutines
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// worker is the main loop for each worker goroutine
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			// Context cancelled, stop processing
			return

		case task, ok := <-wp.tasks:
			if !ok {
				// Channel closed, no more tasks
				return
			}

			// Execute the task
			err := task()

			// Update stats after task completion
			if err != nil {
				wp.mu.Lock()
				wp.stats.FailedTasks++
				continueOnErr := wp.continueOnError
				wp.mu.Unlock()

				// Send error to error channel (non-blocking)
				select {
				case wp.errors <- err:
				default:
				}

				// If not continuing on error, cancel context
				if !continueOnErr {
					wp.cancel()
					return
				}
			} else {
				wp.mu.Lock()
				wp.stats.CompletedTasks++
				wp.mu.Unlock()
			}
		}
	}
}

// Submit adds a task to the worker pool for execution
func (wp *WorkerPool) Submit(task Task) {
	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		panic("cannot submit task to closed worker pool")
	}

	if !wp.started {
		wp.mu.Unlock()
		panic("cannot submit task to worker pool that hasn't been started")
	}

	wp.stats.TotalTasks++
	ctx := wp.ctx
	wp.mu.Unlock()

	// Send task without holding lock to avoid deadlock
	select {
	case wp.tasks <- task:
		// Task submitted successfully
	case <-ctx.Done():
		// Context cancelled, cannot submit
		panic("cannot submit task: context cancelled")
	}
}

// Wait closes the task queue and waits for all workers to finish
// Returns the first error encountered, if any
func (wp *WorkerPool) Wait() error {
	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		return fmt.Errorf("worker pool already closed")
	}
	wp.closed = true
	wp.mu.Unlock()

	// Close the task channel to signal no more tasks
	close(wp.tasks)

	// Wait for all workers to finish
	wp.wg.Wait()

	// Calculate final duration
	wp.mu.Lock()
	wp.stats.Duration = time.Since(wp.startTime)
	wp.mu.Unlock()

	// Check for errors
	select {
	case err := <-wp.errors:
		return err
	default:
	}

	// Check context error
	if wp.ctx.Err() != nil {
		return wp.ctx.Err()
	}

	return nil
}

// Stats returns the current statistics of the worker pool
func (wp *WorkerPool) Stats() PoolStats {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.stats
}

// WorkerCount returns the number of workers in the pool
func (wp *WorkerPool) WorkerCount() int {
	return wp.workers
}
