package glock

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/petermattis/goid"
)

type Mutex struct {
	sync.Mutex

	owner           atomic.Int64 // owner goroutine of the lock, 0 for none
	reentranceCount int64        // count of reentrances in the owner goroutine
}

// Lock and reports whether a waiting occurred
func (me *Mutex) Lock() (waited bool) {
	gid := goid.Get()

	if me.owner.Load() == gid {
		me.reentranceCount++
		return false
	}

	waited = me.owner.Load() == 0
	me.Mutex.Lock()
	me.owner.Store(gid)
	return waited
}

// TryLock only locks successfully when a waiting is not needed. This method is always non-blocking
func (me *Mutex) TryLock() (locked bool) {
	gid := goid.Get()

	if me.owner.Load() == gid { // already owned
		me.reentranceCount++
		return true
	}

	if me.owner.Load() != 0 { // owned by another goroutine, failing the try
		return false
	}

	locked = me.Mutex.TryLock()
	if locked {
		me.owner.Store(gid)
	}
	return locked
}

// L o r+, r- o U

func (me *Mutex) Unlock() {
	gid := goid.Get()
	owner := me.owner.Load()
	owned := gid == owner

	if !owned {
		panic(`unlocking non-owned GLock`)
	}

	if me.reentranceCount == 0 {
		me.owner.Store(0)
		me.Mutex.Unlock()
	} else if me.reentranceCount > 0 {
		me.reentranceCount--
	} else {
		panic(fmt.Errorf(`reentranceCount < 0`))
	}
}
