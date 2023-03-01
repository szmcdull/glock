package glock

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/petermattis/goid"
)

type GLock struct {
	sync.Mutex

	// operations in the owner goroutine does not need synchronizing
	owner           int64 // owner goroutine of the lock, 0 for none
	lockCount       int64 // current lock count
	reentranceCount int64 // count of reentrances in the owner goroutine
}

// Lock and returns true if waited
func (me *GLock) Lock() bool {
	gid := goid.Get()
	lockCount := atomic.AddInt64(&me.lockCount, 1)
	if lockCount == 1 { // first acquire, current goroutine becomes the owner
		me.owner = gid
		me.Mutex.Lock()
		me.reentranceCount++
		return false
	} else if lockCount > 0 {
		if me.owner != gid { // locked in another goroutine, wait to acquire
			me.Mutex.Lock() // wait
			me.owner = gid  // acquired
			me.reentranceCount++
			return true
		}

		// locked in the same goroutine
		me.reentranceCount++
		return false
	}

	panic(fmt.Errorf(`invalid lock count %d`, lockCount))
}

func (me *GLock) Unlock() {
	gid := goid.Get()
	owner := me.owner
	owned := gid == owner

	if owned {
		me.reentranceCount--
	} else {
		panic(`unlocking non-owned GLock`)
	}

	if me.lockCount == 1 {
		me.owner = 0
	}
	atomic.AddInt64(&me.lockCount, -1)

	if me.reentranceCount == 0 {
		me.Mutex.Unlock()
	}

}
