package checker

import (
	"fmt"
	"net"
	"time"

	"upex-wallet/wallet-base/service"
)

type AliveChecker struct {
	service.SimpleWorker
	name    string
	address string
}

func NewAliveChecker(name, address string) *AliveChecker {
	return &AliveChecker{
		name:    name,
		address: address,
	}
}

func (c *AliveChecker) Name() string {
	return "AliveChecker of " + c.name
}

func (c *AliveChecker) Work() {
	conn, err := net.DialTimeout("tcp", c.address, time.Second*5)
	if err != nil {
		body := fmt.Sprintf("service %s (%s) losing contact at %v.", c.name, c.address, time.Now().Format("2006-01-02 15:04:05"))
		warn(body)
		return
	}
	conn.Close()
}
