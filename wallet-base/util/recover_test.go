package util

import (
	"fmt"
	"testing"
	"time"
)

func TestRecover(t *testing.T) {
	WithRecover("test-WithRecover", func() {
		panic("panic here!")
	}, func(err error) {
		fmt.Println("handle err:", err)
	})

	Go("test-Go", func() {
		panic("and panic here!")
	}, func(err error) {
		fmt.Println("handle err:", err)
	})

	time.Sleep(time.Second)
}
