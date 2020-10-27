package v2

import (
	"runtime"

	"upex-wallet/wallet-base/util"
)

func tryGC(taskIdx int) {
	if taskIdx%50000 == 0 {
		runtime.GC()
	}
}

func traceProgress(tag string, num, total int) {
	util.TraceProgress(tag, num, 100000, total)
}
