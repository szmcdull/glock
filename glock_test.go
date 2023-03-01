package glock

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestReentrance(t *testing.T) {
	ch := make(chan struct{})

	go func() {
		l := GLock{}
		l.Lock()
		l.Lock()
		l.Unlock()
		l.Unlock()
		if l.owner != 0 {
			t.Error(`owner not cleared`)
		}

		if l.reentranceCount != 0 {
			t.Error(`reentranceCount not cleared`)
		}

		if l.lockCount != 0 {
			t.Error(`lockCount not cleared`)
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
		l := GLock{}
		if l.TryLock() != true {
			t.Fail()
		}
		if l.TryLock() != true {
			t.Error(`2`)
		}
		l.Unlock()
		l.Unlock()
		if l.owner != 0 {
			t.Error(`owner not cleared`)
		}

		if l.reentranceCount != 0 {
			t.Error(`reentranceCount not cleared`)
		}

		if l.lockCount != 0 {
			t.Error(`lockCount not cleared`)
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
	l := GLock{}
	l.Lock()
	step := 0

	go func() {
		l.Lock()
		step = 1
		l.Unlock()
	}()

	time.Sleep(time.Millisecond * 100)
	if step != 0 {
		t.Error(`not locked`)
	}

	l.Unlock()
	time.Sleep(time.Millisecond * 100)
	if step != 1 {
		t.Error(`not released`)
	}

	if l.owner != 0 {
		t.Error(`owner not cleared`)
	}

	if l.reentranceCount != 0 {
		t.Error(`reentranceCount not cleared`)
	}

	if l.lockCount != 0 {
		t.Error(`lockCount not cleared`)
	}
}

func TestLockTry(t *testing.T) {
	l := GLock{}
	if l.TryLock() != true {
		t.Error(`1`)
	}
	step := 0

	go func() {
		if l.TryLock() != false {
			t.Error(`2`)
		}
		l.Lock()
		step = 1
		l.Unlock()
	}()

	time.Sleep(time.Millisecond * 100)
	if step != 0 {
		t.Error(`not locked`)
	}

	l.Unlock()
	time.Sleep(time.Millisecond * 100)
	if step != 1 {
		t.Error(`not released`)
	}

	if l.owner != 0 {
		t.Error(`owner not cleared`)
	}

	if l.reentranceCount != 0 {
		t.Error(`reentranceCount not cleared`)
	}

	if l.lockCount != 0 {
		t.Error(`lockCount not cleared`)
	}
}

func TestConcurrent(t *testing.T) {
	l := GLock{}
	count := 1000000
	wg := sync.WaitGroup{}
	wg.Add(count)
	counted := 0

	for i := 0; i < count; i++ {
		go func() {
			c := rand.Intn(16) + 1
			for j := 0; j < c; j++ {
				l.Lock()
			}
			counted++
			for j := 0; j < c; j++ {
				l.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()
	if /*l.owner != 0 || */ l.reentranceCount != 0 || l.lockCount != 0 {
		t.Fail()
	}
	if counted != count {
		t.Fail()
	}
}

func TestConcurrentTry(t *testing.T) {
	l := GLock{}
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
	if l.owner != 0 || l.reentranceCount != 0 || l.lockCount != 0 {
		t.Fail()
	}
	if counted != int64(count) {
		t.Fail()
	}
}
