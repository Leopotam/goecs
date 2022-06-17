// ----------------------------------------------------------------------------
// The Proprietary or MIT-Red License
// Copyright (c) 2012-2022 Leopotam <leopotam@yandex.ru>
// ----------------------------------------------------------------------------

package ecsmt

import (
	"runtime"
	"sync"

	"leopotam.com/go/ecs"
)

type worker struct {
	id          int
	proc        func(worker *worker)
	workPresent chan struct{}
	workDone    chan struct{}
	entities    []int
	from        int
	before      int
}

type ITask interface {
	Process(entities []int, from, before int)
}

var workers []*worker
var task ITask
var runSync sync.Mutex

func workerProc(worker *worker) {
	for {
		<-worker.workPresent
		task.Process(worker.entities, worker.from, worker.before)
		worker.entities = nil
		worker.workDone <- struct{}{}
	}
}

func Run(newTask ITask, filter *ecs.Filter, chunkSize int) {
	runSync.Lock()
	count := filter.GetEntitiesCount()
	if count <= 0 || chunkSize <= 0 {
		runSync.Unlock()
		return
	}
	maxWorkers := len(workers)
	if maxWorkers == 0 {
		for i := 0; i < runtime.NumCPU(); i++ {
			w := &worker{
				id:          i,
				proc:        workerProc,
				workPresent: make(chan struct{}),
				workDone:    make(chan struct{}),
			}
			workers = append(workers, w)
			maxWorkers++
			go w.proc(w)
		}
	}
	task = newTask
	processed := 0
	jobSize := count / maxWorkers
	entities := filter.GetRawEntities()
	var workersCount int
	if jobSize >= chunkSize {
		workersCount = maxWorkers
	} else {
		workersCount = count / chunkSize
		jobSize = chunkSize
	}
	if workersCount <= 0 {
		workersCount = 1
	}
	for _, v := range workers[:workersCount-1] {
		v.entities = entities
		v.from = processed
		processed += jobSize
		v.before = processed
		v.workPresent <- struct{}{}
	}
	lastWorker := workers[workersCount-1]
	lastWorker.entities = entities
	lastWorker.from = processed
	lastWorker.before = count
	lastWorker.workPresent <- struct{}{}
	for _, v := range workers[:workersCount] {
		<-v.workDone
	}
	task = nil
	runSync.Unlock()
}
