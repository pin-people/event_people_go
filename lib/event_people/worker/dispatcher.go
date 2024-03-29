package worker

import (
	"sync"
)

type Pool struct {
	singleJob         chan Work
	readyPool         chan chan Work //boss says hey i have a new job at my desk workers who available can get it in this way he does not have to ask current status of workers
	workers           []*worker
	dispatcherStopped sync.WaitGroup
	workersStopped    *sync.WaitGroup
	quit              chan bool
}

var mu = &sync.Mutex{}
var workersInUse = 0
var maxWorkersInPool = 0

func NewWorkerPool(maxWorkers int) *Pool {
	maxWorkersInPool = maxWorkers

	workersStopped := sync.WaitGroup{}

	readyPool := make(chan chan Work, maxWorkers)
	workers := make([]*worker, maxWorkers, maxWorkers)

	// create workers
	for i := 0; i < maxWorkers; i++ {
		workers[i] = NewWorker(i+1, readyPool, &workersStopped)
	}

	return &Pool{
		singleJob:         make(chan Work),
		readyPool:         readyPool,
		workers:           workers,
		dispatcherStopped: sync.WaitGroup{},
		workersStopped:    &workersStopped,
		quit:              make(chan bool),
	}
}

func (q *Pool) Start() {
	//tell workers to get ready
	for i := 0; i < len(q.workers); i++ {
		q.workers[i].Start()
	}
	// open factory
	go q.dispatch()
}

func (q *Pool) Stop() {
	q.quit <- true
	q.dispatcherStopped.Wait()
}

func (q *Pool) dispatch() {
	//open factory gate
	q.dispatcherStopped.Add(1)
	for {
		select {
		case job := <-q.singleJob:
			workerXChannel := <-q.readyPool //free worker x founded
			workerXChannel <- job           // here is your job worker x
		case <-q.quit:
			// free all workers
			for i := 0; i < len(q.workers); i++ {
				q.workers[i].Stop()
			}
			// wait for all workers to finish their job
			q.workersStopped.Wait()
			//close factory gate
			q.dispatcherStopped.Done()
			return
		}
	}
}

func (q *Pool) GetWorkersInUse() int {
	return workersInUse
}

func (q *Pool) GetMaxWorkersInPool() int {
	return maxWorkersInPool
}

func (q *Pool) AddWorkerCount() {
	// add new worker to pool
	workersInUse++
}

func (q *Pool) RemoveWorkerCount() {
	// remove worker from pool
	workersInUse--
}

func (q *Pool) IsWorkerAvailable() bool {
	return workersInUse < maxWorkersInPool
}

func (q *Pool) GetWorkerStatus() (int, int) {
	return workersInUse, maxWorkersInPool
}

/*This is blocking if all workers are busy*/
func (q *Pool) Submit(job Work) {
	// daily - fill the board with new works
	q.singleJob <- job
}
