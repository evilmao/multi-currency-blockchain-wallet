package rpc

import (
	"fmt"
	"sync"
	"time"

	"upex-wallet/wallet-base/util"
)

const (
	maxBatchNumber = 20
)

func BytesBlockGetter(getter func(uint64) ([]byte, error)) func(uint64) (interface{}, error) {
	return func(height uint64) (interface{}, error) {
		return getter(height)
	}
}

type ErrHeightOver struct {
	bestHeight uint64
}

func NewErrHeightOver(bestHeight uint64) *ErrHeightOver {
	return &ErrHeightOver{bestHeight}
}

func (e *ErrHeightOver) Error() string {
	return fmt.Sprintf("height is greater than the best height of %d", e.bestHeight)
}

type BlockCache struct {
	sync.RWMutex

	bestHeightGetter func() (uint64, error)
	blockGetter      func(uint64) (interface{}, error)

	cacheIndex map[uint64]interface{}
}

func NewBlockCache(
	bestHeightGetter func() (uint64, error),
	blockGetter func(uint64) (interface{}, error)) *BlockCache {

	c := &BlockCache{}
	c.Reset()

	c.bestHeightGetter = func() (uint64, error) {
		var bestHeight uint64
		err := util.Try(3, func(int) error {
			h, err := bestHeightGetter()
			if err != nil {
				return err
			}

			bestHeight = h
			return nil
		})
		return bestHeight, err
	}

	c.blockGetter = func(height uint64) (interface{}, error) {
		var data interface{}
		err := util.Try(3, func(int) error {
			d, err := blockGetter(height)
			if err != nil {
				return err
			}

			data = d
			return nil
		})
		return data, err
	}

	return c
}

func (c *BlockCache) Get(height uint64) (interface{}, error) {
	if data, ok := c.get(height); ok {
		return data, nil
	}

	bestHeight, err := c.bestHeightGetter()
	if err != nil {
		return nil, fmt.Errorf("get best height failed, %v", err)
	}

	if bestHeight < height {
		return nil, NewErrHeightOver(bestHeight)
	}

	count := bestHeight - height + 1
	if count > maxBatchNumber {
		count = maxBatchNumber
	}

	return c.batchGet(height, count)
}

func (c *BlockCache) batchGet(fromHeight, count uint64) (interface{}, error) {
	c.Reset()

	var wg sync.WaitGroup
	for i := uint64(0); i < count; i++ {
		wg.Add(1)

		go func(height uint64) {
			defer wg.Done()

			data, err := c.blockGetter(height)
			if err != nil {
				return
			}

			c.add(height, data)
		}(fromHeight + i)

		time.Sleep(time.Millisecond * 10)
	}
	wg.Wait()

	if data, ok := c.get(fromHeight); ok {
		return data, nil
	}

	return nil, fmt.Errorf("get block at height %d failed", fromHeight)
}

func (c *BlockCache) get(height uint64) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	data, ok := c.cacheIndex[height]
	return data, ok
}

func (c *BlockCache) add(height uint64, data interface{}) {
	c.Lock()
	defer c.Unlock()

	c.cacheIndex[height] = data
}

func (c *BlockCache) Reset() {
	c.Lock()
	defer c.Unlock()

	c.cacheIndex = make(map[uint64]interface{}, maxBatchNumber)
}
