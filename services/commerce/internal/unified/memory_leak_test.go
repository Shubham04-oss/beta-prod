package unified

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/google/uuid"
)

func TestKeyMutex_MemoryLeak(t *testing.T) {
	k := &keyMutex{}

	// Record initial memory
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Create 1 million locks
	for i := 0; i < 1000000; i++ {
		key := uuid.New().String()
		unlock := k.Lock(key)
		unlock()
	}

	// Record final memory
	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocBytes := int64(m2.Alloc) - int64(m1.Alloc)
	fmt.Printf("Memory allocated for 1M locks: %d MB\n", allocBytes/1024/1024)

	if allocBytes > 10*1024*1024 {
		t.Errorf("Memory leak detected: %d bytes allocated", allocBytes)
	}
}
