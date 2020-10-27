package util

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type AtomicBool int32

func (b *AtomicBool) Set(is bool) {
	if is {
		atomic.StoreInt32((*int32)(b), 1)
	} else {
		atomic.StoreInt32((*int32)(b), 0)
	}
}

func (b *AtomicBool) Is() bool {
	return atomic.LoadInt32((*int32)(b)) != 0
}

type BatchWork func(idx int) (interface{}, error)
type BatchGather func(idx int, data interface{}) error

// BatchDo does batch, gather is called in the same goroutine as BatchDo in order.
func BatchDo(count int, work BatchWork, gather BatchGather) error {
	return NewBatch(count, work, gather).Do()
}

type Batch struct {
	count  int
	work   BatchWork
	gather BatchGather
}

func NewBatch(count int, work BatchWork, gather BatchGather) *Batch {
	return &Batch{
		count:  count,
		work:   work,
		gather: gather,
	}
}

func (b *Batch) Do() error {
	if b.count <= 0 {
		return nil
	}

	var (
		goroutineNum = runtime.NumCPU() * 8
		groupNum     = (b.count-1)/goroutineNum + 1

		datas = make([]interface{}, goroutineNum)
		errs  = make([]error, goroutineNum)
	)

	for i := 0; i < groupNum; i++ {
		var (
			start = i * goroutineNum
			wg    sync.WaitGroup
		)
		for idx := start; idx < (start+goroutineNum) && idx < b.count; idx++ {
			wg.Add(1)

			workIdx := idx
			dataIdx := idx - start
			Go("Batch.do", func() {
				b.do(&wg, workIdx, dataIdx, datas, errs)
			}, nil)
		}
		wg.Wait()

		for _, err := range errs {
			if err != nil {
				return err
			}
		}

		if b.gather != nil {
			dataNum := goroutineNum
			if (start + dataNum) > b.count {
				dataNum = b.count - start
			}

			for i, data := range datas[:dataNum] {
				err := b.gather(start+i, data)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (b *Batch) do(wg *sync.WaitGroup, workIdx, dataIdx int, datas []interface{}, errs []error) {
	defer wg.Done()

	data, err := b.work(workIdx)
	if err != nil {
		errs[dataIdx] = err
	} else {
		datas[dataIdx] = data
	}
}
