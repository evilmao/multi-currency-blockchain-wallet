package email

import (
	"github.com/go-gomail/gomail"
)

// SimpleSend sends a simple email.
func SimpleSend(host string, port int, sender, pass string, receivers []string, title, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", receivers...)
	m.SetHeader("Subject", title)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(host, port, sender, pass)
	return d.DialAndSend(m)
}
