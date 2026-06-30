package glock

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestReentrance(t *testing.T) {
	ch := make(chan struct{})

	go func() {
		l := Mutex{}
		l.Lock()
		l.Lock()
		l.Unlock()
		l.Unlock()
		if l.owner.Load() != 0 {
			t.Error(`owner not cleared`)
		}

		if l.reentranceCount != 0 {
			t.Error(`reentranceCount not cleared`)
		}

		ch <- struct{}{}
	}()

	timer := time.NewTimer(time.Second)
	select {
	case <-timer.C:
		t.Error(`timed out`)
	case <-ch:
	}
}

func TestReentranceTry(t *testing.T) {
	ch := make(chan struct{})

	go func() {
		l := Mutex{}
		if l.TryLock() != true {
			t.Fail()
		}
		if l.TryLock() != true {
			t.Error(`2`)
		}
		l.Unlock()
		l.Unlock()
		if l.owner.Load() != 0 {
			t.Error(`owner not cleared`)
		}

		if l.reentranceCount != 0 {
			t.Error(`reentranceCount not cleared`)
		}

		ch <- struct{}{}
	}()

	timer := time.NewTimer(time.Second)
	select {
	case <-timer.C:
		t.Fail() // timed out
	case <-ch:
	}
}

func TestLock(t *testing.T) {
	l := Mutex{}
	l.Lock()
	var step atomic.Int32
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		l.Lock()
		step.Store(1)
		l.Unlock()
	}()

	time.Sleep(time.Millisecond * 100)
	if step.Load() != 0 {
		t.Error(`not locked`)
	}

	l.Unlock()
	wg.Wait()
	if step.Load() != 1 {
		t.Error(`not released`)
	}

	if l.owner.Load() != 0 {
		t.Error(`owner not cleared`)
	}

	if l.reentranceCount != 0 {
		t.Error(`reentranceCount not cleared`)
	}
}

func TestLockTry(t *testing.T) {
	l := Mutex{}
	if l.TryLock() != true {
		t.Error(`1`)
	}
	var step atomic.Int32
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if l.TryLock() != false {
			t.Error(`2`)
		}
		l.Lock()
		step.Store(1)
		l.Unlock()
	}()

	time.Sleep(time.Millisecond * 100)
	if step.Load() != 0 {
		t.Error(`not locked`)
	}

	l.Unlock()
	wg.Wait()
	if step.Load() != 1 {
		t.Error(`not released`)
	}

	if l.owner.Load() != 0 {
		t.Error(`owner not cleared`)
	}

	if l.reentranceCount != 0 {
		t.Error(`reentranceCount not cleared`)
	}
}

func TestConcurrent(t *testing.T) {
	l := Mutex{}
	count := 1000000
	wg := sync.WaitGroup{}
	wg.Add(count)
	var counted atomic.Int64

	for i := 0; i < count; i++ {
		go func() {
			c := rand.Intn(16) + 1
			for j := 0; j < c; j++ {
				l.Lock()
			}
			counted.Add(1)
			for j := 0; j < c; j++ {
				l.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()
	if l.owner.Load() != 0 || l.reentranceCount != 0 {
		t.Fail()
	}
	if counted.Load() != int64(count) {
		t.Fail()
	}
}

func TestConcurrentTry(t *testing.T) {
	l := Mutex{}
	count := 1000000
	wg := sync.WaitGroup{}
	wg.Add(count)
	counted := int64(0)

	for i := 0; i < count; i++ {
		go func() {
			c := rand.Intn(16) + 1
			c2 := 0
			added := false
			for j := 0; j < c; j++ {
				if l.TryLock() {
					c2++
				}
			}
			if c2 == 0 {
				l.Lock()
				c2++
				added = true
			}
			counted++
			for j := 0; j < c2; j++ {
				l.Unlock()
			}
			if added {

			}
			wg.Done()
		}()
	}

	wg.Wait()
	if l.owner.Load() != 0 || l.reentranceCount != 0 {
		t.Fail()
	}
	if counted != int64(count) {
		t.Fail()
	}
}
