package alarm

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/jordan-wright/email"

	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"
)

var (
	emailTitle = "[UPEX 告警]"
)

func sendEmailByHTML(cfg *config.Config, task *models.Tx, errMsg string) (err error) {

	var (
		fromAddress     = cfg.EmailCfg.From
		toAddress       = cfg.EmailCfg.To
		password        = cfg.EmailCfg.Pwd
		host            = cfg.EmailCfg.Host
		port            = cfg.EmailCfg.Port

		server  = fmt.Sprintf("%s:%s", host, port)
		txType  = models.TxTypeName(task.TxType)
		errTime = task.CreatedAt.Format("2006-01-02 15:04:05")

		txAddress = task.Address
		currency  = strings.ToUpper(task.Symbol)
		txTransID = task.TransID

		auth = smtp.PlainAuth("", fromAddress, password, host)
	)

	e := email.NewEmail()
	// sender
	e.From = fromAddress
	// to users
	e.To = []string{toAddress}
	// email title
	e.Subject = emailTitle
	// Parse html template
	t, err := template.ParseFiles("email-template.html")
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)

	err = t.Execute(body, struct {
		TxType          string
		ErrorDetail     string
		TimeDate        string
		Currency        string
		TxAddress       string
		TxID            string
	}{
		TxType:          txType,
		ErrorDetail:     errMsg,
		TimeDate:        errTime,
		Currency:        currency,
		TxAddress:       txAddress,
		TxID:            txTransID,
	})

	if err != nil {
		fmt.Println(err)
	}

	e.HTML = body.Bytes()

	return e.SendWithTLS(server, auth, &tls.Config{ServerName: host})

}

func SendEmail(cfg *config.Config, task *models.Tx, err error, msg string) {

	if err == nil {
		return
	}

	// update error to catch
	ok := Update(cfg, task, err)

	if ok {
		err := sendEmailByHTML(cfg, task, msg)
		if err != nil {
			log.Errorf("send email error,%v", err)
		}
	}
}

// alarm func, support email warning
func SendEmailByText(cfg *config.Config, content string) {

	var (
		fromAddress = cfg.EmailCfg.From
		toAddress   = cfg.EmailCfg.To
		password    = cfg.EmailCfg.Pwd
		server      = fmt.Sprintf("%s:%s", cfg.EmailCfg.Host, cfg.EmailCfg.Port)
		conn        = &tls.Conn{}
		Time        = time.Now().Format("2006-01-02 15:04:05")
	)

	var (
		from = mail.Address{Name: "UPEX", Address: fromAddress}
		to   = mail.Address{Address: toAddress}

		// Setup headers
		headers = make(map[string]string)
		body    = fmt.Sprintf(`
		
		你好，Administrator ：

		  - 详情: %s
		  - 时间：%s

		请知悉!!!
			`, content, Time)
	)

	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = emailTitle

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP_Server
	host, _, _ := net.SplitHostPort(server)

	auth := smtp.PlainAuth("", fromAddress, password, host)

	// TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", server, tlsConfig)

	if err != nil {
		log.Panic(err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Panic(err)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		log.Panic(err)
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return
	}

	if err = c.Rcpt(to.Address); err != nil {
		return
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return
	}

	err = w.Close()
	if err != nil {
		return
	}

	// send data
	err = c.Quit()
	if err != nil {
		return
	}

	log.Info("send email success!")
	return
}
