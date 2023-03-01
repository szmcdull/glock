package glock

import (
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

// o l+ L r+,
// Lock and returns true if waited
func (me *GLock) Lock() bool {
	gid := goid.Get()

	// if me.owner == gid {
	// 	me.reentranceCount++
	// 	return false
	// }

	// me.Mutex.Lock()
	// result := me.owner == 0
	// me.owner = gid
	// return result

	if atomic.CompareAndSwapInt64(&me.owner, 0, gid) { // first acquire, current goroutine becomes the owner
		atomic.AddInt64(&me.lockCount, 1)
		// if lockCount != 1 {
		// 	panic(fmt.Errorf(`wrong lockCount %d, 1 expected`, lockCount))
		// }
		//if lockCount == 1 {
		me.Mutex.Lock()
		//}
		me.reentranceCount++
		return false
	} else { // locked in another goroutine, wait to acquire
		atomic.AddInt64(&me.lockCount, 1)
		// if lockCount <= 1 {
		// 	panic(fmt.Errorf(`wrong lockCount %d, expecting > 1`, lockCount))
		// }
		if me.owner == gid {
			me.reentranceCount++
			return false
		}

		me.Mutex.Lock() // wait
		me.owner = gid  // acquired
		me.reentranceCount++
		return true
	}
}

// o l+ L r+,
// TryLock only locks successfully when a waiting is not needed. This method is always non-blocking
func (me *GLock) TryLock() bool {
	gid := goid.Get()
	if me.owner == gid { // already owned
		me.Lock()
		return true
	}
	if me.owner != 0 { // owned by another goroutine, failing the try
		return false
	}

	// me.owner == 0, not owned by any goroutine yet. so try to acquire the lock now
	if atomic.CompareAndSwapInt64(&me.owner, 0, gid) { // first acquire, current goroutine becomes the owner
		atomic.AddInt64(&me.lockCount, 1)
		me.owner = gid
		me.Mutex.Lock()
		me.reentranceCount++
		return true
	} else { // lock grabbed by another goroutine first
		return false
	}
}

// o l+ L r+, r- U l- o

func (me *GLock) Unlock() {
	gid := goid.Get()
	owner := me.owner
	owned := gid == owner

	if owned {
		me.reentranceCount--
	} else {
		panic(`unlocking non-owned GLock`)
	}

	// // lockCount may be increased by other goroutines (before waiting for this lock)
	// // so me.owner may not be cleared after a full unlocking. but it doesn't matter, because lockCount will be 0
	// if me.lockCount == 1 {
	// 	me.owner = 0
	// }

	if me.reentranceCount == 0 {
		me.Mutex.Unlock()
	}

	lockCount := atomic.AddInt64(&me.lockCount, -1)

	if lockCount == 0 {
		me.owner = 0
	}
}
