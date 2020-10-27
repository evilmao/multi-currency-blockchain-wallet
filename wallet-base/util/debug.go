package util

import (
	"fmt"
	"runtime"
)

func TraceMemStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("Alloc: %s, HeapIdle: %s, HeapReleased: %s\n",
		ByteCountBinary(int64(ms.Alloc)), ByteCountBinary(int64(ms.HeapIdle)), ByteCountBinary(int64(ms.HeapReleased)))
}
