package glock

import (
	"fmt"
	"sync"

	"github.com/petermattis/goid"
)

type GLock struct {
	sync.Mutex

	// operations in the owner goroutine does not need synchronizing
	owner           int64 // owner goroutine of the lock, 0 for none
	lockCount       int64 // current lock count
	reentranceCount int64 // count of reentrances in the owner goroutine
}

// L o r+
// Lock and returns true if waited
func (me *GLock) Lock() (waited bool) {
	gid := goid.Get()

	if me.owner == gid {
		me.reentranceCount++
		return false
	}

	waited = me.owner == 0
	me.Mutex.Lock()
	me.owner = gid
	return waited
}

// L o r+
// TryLock only locks successfully when a waiting is not needed. This method is always non-blocking
func (me *GLock) TryLock() (locked bool) {
	gid := goid.Get()
	if me.owner == gid { // already owned
		me.reentranceCount++
		return true
	}
	if me.owner != 0 { // owned by another goroutine, failing the try
		return false
	}

	locked = me.Mutex.TryLock()
	if locked {
		me.owner = gid
	}
	return locked
}

// L o r+, r- o U

func (me *GLock) Unlock() {
	gid := goid.Get()
	owner := me.owner
	owned := gid == owner

	if !owned {
		panic(`unlocking non-owned GLock`)
	}

	if me.reentranceCount == 0 {
		me.owner = 0
		me.Mutex.Unlock()
	} else if me.reentranceCount > 0 {
		me.reentranceCount--
	} else {
		panic(fmt.Errorf(`reentranceCount < 0`))
	}

	// lockCount := atomic.AddInt64(&me.lockCount, -1)

	// if lockCount == 0 {
	// 	me.owner = 0
	// }
}
