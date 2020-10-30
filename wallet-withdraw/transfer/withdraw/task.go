package withdraw

import (
	"fmt"
	"strings"
	"time"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/blockchain"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/withdraw/transfer/config"
	"upex-wallet/wallet-withdraw/base/models"

	"github.com/shopspring/decimal"
)

type taskProducer struct {
	cfg    *config.Config
	exAPI  *api.ExAPI
	taskCh chan *models.Tx

	lastUpdateTime time.Time
}

func newTaskProducer(cfg *config.Config, exAPI *api.ExAPI) *taskProducer {
	return &taskProducer{
		cfg:    cfg,
		exAPI:  exAPI,
		taskCh: make(chan *models.Tx, 10000),
	}
}

func (p *taskProducer) Name() string {
	return "taskProducer"
}

func (p *taskProducer) Init() error {
	return nil
}

func (p *taskProducer) Work() {
	if len(p.taskCh) > 0 {
		return
	}

	p.produceFromAPIs()
	p.produceFromDB()
}

func (p *taskProducer) Destroy() {
	close(p.taskCh)
}

func (p *taskProducer) next() (*models.Tx, bool) {
	var (
		t  *models.Tx
		ok bool
	)

	select {
	case t, ok = <-p.taskCh:
		p.lastUpdateTime = time.Now()
	default:
		if time.Now().Sub(p.lastUpdateTime) >= time.Minute {
			log.Infof("%s, wait for tasks...", p.Name())
			p.lastUpdateTime = time.Now()
		}
	}
	return t, ok
}

func (p *taskProducer) produceFromAPIs() {

	p.produceFromAPI()
}

func (p *taskProducer) produceFromAPI() {

	var data = make(map[string]interface{})
	data["symbol"] = strings.ToLower(p.cfg.Currency)
	data["app_id"] = p.cfg.BrokerAccessKey
	data["timestamp"] = time.Now().Unix()

	// get withdraws tasks form broker
	resp, _, err := p.exAPI.GetWithdraws(data)
	if err != nil || resp == nil {
		if err != nil {
			log.Errorf("%s, get api withdraw tasks failed, %v", p.Name(), err)
		}
		return
	}

	datas := resp.([]interface{})
	if len(datas) == 0 {
		return
	}

	for _, data := range datas {
		var (
			d         = data.(map[string]interface{})
			id        = d["trans_id"].(float64)
			amount    = d["amount"].(float64)
			addressTo = d["address_to"].(string)
		)

		if addressTo == "" || amount <= 0 {
			continue
		}

		var task models.Tx
		task.TransID = fmt.Sprintf("%.f", id) // from broker
		task.SequenceID = util.HashString32([]byte(task.TransID))
		task.Address = addressTo
		task.Symbol = strings.ToLower(p.cfg.Currency)
		task.TxType = models.TxTypeWithdraw
		task.Amount = decimal.NewFromFloat(amount)
		task.Fees = decimal.NewFromFloat(d["fee"].(float64))
		p.taskCh <- &task
	}
}

func (p *taskProducer) produceFromDB() {
	txs := models.GetUnfinishedWithdraws()
	for _, tx := range txs {
		p.taskCh <- tx
	}
}

func (p *taskProducer) acceptBlockchainCurrency(blockchainName, currency string) bool {
	name, canEmpty := blockchain.CorrectName(currency, p.cfg.Currency)
	if blockchainName == name {
		return true
	}

	if len(blockchainName) == 0 && canEmpty {
		return true
	}

	return false
}
