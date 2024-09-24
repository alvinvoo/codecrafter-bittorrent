package worker

import (
	"context"
	"sync"

	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/util"
)

// Job represents a interface of job that can be enqueued into a dispatcher.
type Job interface{}

type Worker interface {
	Work(j Job)
}

// Dispatcher represents a job dispatcher.
type Dispatcher struct {
	sem       chan struct{} // semaphore
	jobBuffer chan Job
	worker    Worker
	wg        sync.WaitGroup
}

// NewDispatcher will create a new instance of job dispatcher.
// maxWorkers means the maximum number of goroutines that can work concurrently.
// buffers means the maximum size of the queue.
func NewDispatcher(worker Worker, maxWorkers int, buffers int) *Dispatcher {
	return &Dispatcher{
		// Restrict the number of goroutine using buffered channel (as counting semaphor)
		sem:       make(chan struct{}, maxWorkers),
		jobBuffer: make(chan Job, buffers),
		worker:    worker,
	}
}

// Start starts a dispatcher.
// This dispatcher will stops when it receive a value from `ctx.Done`.
func (d *Dispatcher) Start(ctx context.Context) {
	d.wg.Add(1)
	util.DebugLog("starting dispatcher loop")
	go d.loop(ctx)
}

// Wait blocks until the dispatcher stops.
func (d *Dispatcher) Wait() {
	d.wg.Wait()
}

// Add enqueues a job into the queue.
// If the number of enqueued jobs has already reached to the maximum size,
// this will block until the other job has finish and the queue has space to accept a new job.
func (d *Dispatcher) Add(job Job) {
	d.jobBuffer <- job
}

func (d *Dispatcher) Stop() {
	d.wg.Done()
}

func (d *Dispatcher) loop(ctx context.Context) {
	var wg sync.WaitGroup

	// IMPT: Monitor completion of all jobs
	go func() {
		util.DebugLog("monitoring jobs")
		wg.Wait()
		util.DebugLog("stopping loop")
		d.Stop() // Signal the dispatcher to stop when all jobs are done
	}()

	// the for-loop will run first if there's blocking operations
	// only when it is unblocked or have waiting periods, the go routines will run
	// but in theory, go routines are supposed to run concurrently
	// so to avoid any potential race conditions, the wait group actually needs to be incremented
	// even before this function
Loop:
	for {
		select {
		case <-ctx.Done():
			// block until all the jobs finishes
			wg.Wait()
			break Loop // need to specify a label to break the `for` loop; else only breaking `select` loop
		case job := <-d.jobBuffer:
			// Increment the waitgroup
			util.DebugLog("incrementing job waitgroup")
			wg.Add(1)
			// Decrement a semaphore count
			// Will block if semaphore channel buffer is full
			d.sem <- struct{}{}
			go func(job Job) {
				defer wg.Done()
				// After the job finished, increment a semaphore count
				defer func() { <-d.sem }()
				d.worker.Work(job)
			}(job)
		}
	}
}
