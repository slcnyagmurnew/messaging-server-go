package scheduler

import (
	"context"
	"fmt"
	"messaging-server/internal/logger"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// Scheduler runs a provided job func every interval
// Start and Stop it on demand
type Scheduler struct {
	interval time.Duration
	job      func(ctx context.Context)

	cancel func()         // to stop the loop
	wg     sync.WaitGroup // wait for running jobs
	done   chan struct{}  // closed when the loop exits
}

// New creates a Scheduler that will fire job every interval
func New(job func(ctx context.Context)) (*Scheduler, error) {
	// set interval value from environment
	var interval time.Duration

	if sec, err := strconv.Atoi(os.Getenv("SCHEDULER_INTERVAL")); err == nil {
		interval = time.Duration(sec) * time.Second
	} else {
		// default 1 hour
		interval = time.Duration(120) * time.Second
	}

	if interval <= 0 {
		err := fmt.Errorf("interval must be > 0, got %s", interval) // catch interval error
		return nil, err
	}

	return &Scheduler{
		interval: interval,
		job:      job,
		wg:       sync.WaitGroup{},
		done:     make(chan struct{}),
	}, nil
}

// Start begins the scheduling loop
func (s *Scheduler) Start() (string, int) {
	if s.cancel != nil {
		logger.Sugar.Info("scheduler already running")
		return "scheduler already running", http.StatusOK // already running
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		defer close(s.done)

		// start ticker with given interval
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// infinite loop until context done
		for {
			select {
			case <-ticker.C: // if context is active, run job concurrently in given interval
				// track job (graceful shutdown)
				s.wg.Add(1)
				go func() {
					defer s.wg.Done()
					s.job(ctx)
				}()
			case <-ctx.Done():
				return
			}
		}
	}()
	logger.Sugar.Info("scheduler started")
	return "scheduler started", http.StatusCreated
}

// Stop signals the scheduler to stop and waits for it to exit.
func (s *Scheduler) Stop() {
	if s.cancel == nil {
		logger.Sugar.Info("scheduler is not running")
		return
	}
	s.cancel()
	<-s.done

	// wait for any jobs that were started to finish
	s.wg.Wait()

	s.cancel = nil
	s.done = make(chan struct{})

	logger.Sugar.Info("scheduler stopped")
	return
}
