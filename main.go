/**
 * @Author: peer
 * @Date: 2021-05-13 23:21:47
 * @LastEditTime: 2021-06-21 00:59:57
 * @Description: file content
 */
package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Task struct {
	ID      uint64
	Handler func(v ...interface{}) interface{}
	Params  []interface{}
}

type Pool struct {
	capacity       uint64
	runningWorkers uint64
	status         int64
	chTask         chan *Task
	sync.Mutex
	panicHandler func(interface{})
}

var ErrInvalidPoolCap = errors.New("invalid pool cap")

const (
	RUNNING = 1
	STOPED  = 0
)

func NewPool(capacity uint64) (*Pool, error) {
	if capacity <= 0 {
		return nil, ErrInvalidPoolCap
	}
	return &Pool{
		capacity: capacity,
		status:   RUNNING,
		chTask:   make(chan *Task, capacity),
	}, nil
}

func (p *Pool) run() {
	p.incRunning()
	go func() {
		defer func() {
			p.decRunning()
			if r := recover(); r != nil {
				if p.panicHandler != nil {
					p.panicHandler(r)
				} else {
					log.Printf("worker panic %s\n", r)
				}
			}
			p.checkWorker()
		}()
		for {
			select {
			case task := <-p.chTask:
				task.Handler(task.Params...)
			default:
				return
			}
		}
	}()
}

func (p *Pool) incRunning() {
	atomic.AddUint64(&p.runningWorkers, 1)
}

func (p *Pool) decRunning() {
	atomic.AddUint64(&p.runningWorkers, ^uint64(0))
}

func (p *Pool) GetRunningWorkers() uint64 {
	return atomic.LoadUint64(&p.runningWorkers)
}

func (p *Pool) GetCap() uint64 {
	return p.capacity
}

func (p *Pool) setStatus(status int64) bool {
	p.Lock()
	defer p.Unlock()
	if p.status == status {
		return false
	}
	p.status = status
	return true
}

var ErrPoolAlreadyClosed = errors.New("pool already closed")

func (p *Pool) Put(task *Task) error {
	p.Lock()
	defer p.Unlock()
	if p.status == STOPED {
		return ErrPoolAlreadyClosed
	}
	// p.incUnFinishTask(1)
	if p.GetRunningWorkers() < p.GetCap() {
		p.run()
	}
	if p.status == RUNNING {
		p.chTask <- task
		fmt.Println("id", task.ID)
	}
	return nil
}

func (p *Pool) Close() {
	p.setStatus(STOPED)
	for len(p.chTask) > 0 {
		time.Sleep(1e6)
	}
	close(p.chTask)
}

func (p *Pool) checkWorker() {
	p.Lock()
	defer p.Unlock()
	if p.runningWorkers == 0 && len(p.chTask) > 0 {
		p.run()
	}
}

func main() {
	pool, err := NewPool(10)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 20; i++ {
		_ = pool.Put(&Task{
			ID: uint64(i),
			Handler: func(v ...interface{}) interface{} {
				fmt.Println(v)
				time.Sleep(2 * time.Second)

				return v
			},
			Params: []interface{}{i},
		})
	}

	for pool.GetRunningWorkers() != 0 {

	}
}
