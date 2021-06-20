/**
 * @Author: peer
 * @Date: 2021-05-14 00:46:36
 * @LastEditTime: 2021-05-14 01:24:35
 * @Description: file content
 */
package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

var wg = sync.WaitGroup{}

var sum int64

func demoTask(v ...interface{}) interface{} {
	defer wg.Done()
	for i := 0; i < 100; i++ {
		atomic.AddInt64(&sum, 1)
	}
	return nil
}

var runtimes = 1000000

func BenchmarkGoroutineTimeLifeSetTimes(b *testing.B) {
	for i := 0; i < runtimes; i++ {
		wg.Add(1)
		go demoTask()
	}
	wg.Wait()
}

func BenchmarkPoolTimeLifeSetTimes(b *testing.B) {
	pool, err := NewPool(20)
	if err != nil {
		b.Error(err)
	}
	task := &Task{
		Handler: demoTask,
	}
	for i := 0; i <= runtimes; i++ {
		wg.Add(1)
		pool.Put(task)
	}
	wg.Wait()
}

func TestPool(t *testing.T) {
	a := 1
	if a == 1 {
		t.Error("test error")
	}
}
