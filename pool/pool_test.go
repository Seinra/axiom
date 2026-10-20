package pool_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/seinra/axiom/pool"
)

const (
	hammerWorkerCount = 32
	hammerRoundCount  = 200
)

func TestZeroValueGetReturnsZeroWithoutPanic(t *testing.T) {
	var spare pool.Pool[[]byte]
	fetched := spare.Get()
	if fetched != nil {
		t.Fatalf("Get = %v, want nil slice", fetched)
	}
}

func TestZeroValuePutThenGetStaysUsable(t *testing.T) {
	// Validity, not identity: the runtime may drop the Put buffer,
	// so only assert the Get buffer is usable.
	var spare pool.Pool[[]byte]
	spare.Put([]byte("transient"))
	fetched := spare.Get()
	fetched = append(fetched, 'x')
	if len(fetched) == 0 {
		t.Fatal("recycled buffer is unusable, want appendable buffer")
	}
}

func TestNilMakerFallsBackToZeroMaker(t *testing.T) {
	pooled := pool.New[[]byte](nil)
	if pooled == nil {
		t.Fatal("New(nil) = nil, want non-nil pool")
	}
	fetched := pooled.Get()
	if fetched != nil {
		t.Fatalf("Get = %v, want nil slice", fetched)
	}
	pooled.Put([]byte("transient"))
	usable := pooled.Get()
	usable = append(usable, 'x')
	if len(usable) == 0 {
		t.Fatal("recycled buffer is unusable, want appendable buffer")
	}
}

func TestMakerRunsAtLeastOnce(t *testing.T) {
	var calls atomic.Int64
	pooled := pool.New(func() []byte {
		calls.Add(1)
		return make([]byte, 0, 8)
	})
	fetched := pooled.Get()
	if calls.Load() < 1 {
		t.Fatal("maker calls = 0, want at least 1")
	}
	if fetched == nil {
		t.Fatal("Get = nil, want usable buffer")
	}
}

func TestRecycledBufferValidityNotIdentity(t *testing.T) {
	// Teachable caution: NEVER assert Get returns the exact Put buffer.
	// The runtime may drop pooled buffers at any GC, so assert validity
	// (a usable buffer) and that the maker ran at least once.
	var calls atomic.Int64
	pooled := pool.New(func() []byte {
		calls.Add(1)
		return make([]byte, 0, 16)
	})
	seeded := pooled.Get()
	pooled.Put(seeded)
	recycled := pooled.Get()
	recycled = append(recycled, 'x')
	if len(recycled) == 0 {
		t.Fatal("recycled buffer is unusable, want appendable buffer")
	}
	if calls.Load() < 1 {
		t.Fatal("maker calls = 0, want at least 1")
	}
}

func TestConcurrentGetPutHammerIsRaceClean(t *testing.T) {
	var calls atomic.Int64
	pooled := pool.New(func() []byte {
		calls.Add(1)
		return make([]byte, 0, 64)
	})
	var waitGroup sync.WaitGroup
	for workerIdx := 0; workerIdx < hammerWorkerCount; workerIdx++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for roundIdx := 0; roundIdx < hammerRoundCount; roundIdx++ {
				fetched := pooled.Get()
				fetched = append(fetched, 'x')
				pooled.Put(fetched[:0])
			}
		}()
	}
	waitGroup.Wait()
	if calls.Load() < 1 {
		t.Fatal("maker calls = 0, want at least 1")
	}
}
