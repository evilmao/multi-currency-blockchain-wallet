package checker

import (
	"upex-wallet/wallet-tools/base/email"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

// Default check interval in minute.
const (
	DefaultCheckInterval = 30
)

var (
	EmailPassword string
	Receivers     []string
)

var (
	host   = "smtp.aliyun.com"
	port   = 465
	sender = "xxx@aliyun.com"

	warn = func(body string) {
		if len(Receivers) > 0 {
			title := "Warning"
			err := email.SimpleSend(host, port, sender, EmailPassword, Receivers, title, body)
			if err != nil {
				log.Errorf("failed to send email (%s, %s), %v", title, body, err)
			}
		} else {
			log.Error(body)
		}
	}
)
