package s3tools

import "sync"

//InitDispatcher create job queue and start dispatcher
func InitDispatcher(dispatchers int, jobs int) *Dispatcher {
	JobQueue = make(chan Job, jobs)
	dispatcher := NewDispatcher(dispatchers)
	dispatcher.Run()
	return dispatcher
}

// DispatcherWaitGroup control dispatcher tasks count
var DispatcherWaitGroup sync.WaitGroup

//Dispatcher handles pool of workers
type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job
	MaxWorkers int
}

//NewDispatcher creates dispatcher with max workers count
func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{WorkerPool: pool, MaxWorkers: maxWorkers}
}

//EnqueueJob adds a job to a job queue
func (d *Dispatcher) EnqueueJob(job Job) {
	DispatcherWaitGroup.Add(1)
	JobQueue <- job
}

//Run dispatcher in a coroutine
func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.MaxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}
