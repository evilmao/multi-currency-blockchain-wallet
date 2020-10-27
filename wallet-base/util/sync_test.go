package util

import (
	"testing"
)

func TestBatch(t *testing.T) {
	BatchDo(10, func(idx int) (interface{}, error) {
		t.Log("work index:", idx)
		return idx * idx, nil
	}, func(idx int, data interface{}) error {
		t.Log("gather:", idx, data)
		return nil
	})
}
