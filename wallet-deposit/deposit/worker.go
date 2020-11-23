package deposit

import (
	"fmt"
	"strconv"
	"time"

	"upex-wallet/wallet-base/currency"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/monitor"
	"upex-wallet/wallet-base/newbitx/misclib/log"
	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-base/util"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/rpc"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

const (
	workerTag = "deposit.worker"

	TxAmountPrecision = 18
)

var (
	// Worker must implement service.Worker.
	_ service.Worker = &Worker{}
)

type Worker struct {
	cfg                    *config.Config
	rpcClient              rpc.RPC
	currentBlock           *rpc.Block
	notifier               *notifier
	notifierSrv            *service.Service
	updateMonitor          *util.UpdateMonitor
	lastLogInsertBlockTime time.Time
}

func New(cfg *config.Config, rpcClient rpc.RPC) *Worker {
	w := &Worker{
		cfg:                    cfg,
		rpcClient:              rpcClient,
		lastLogInsertBlockTime: time.Now(),
	}

	// notifier
	if !cfg.IgnoreNotifyAudit {
		w.notifier = newNotifier(cfg, rpcClient)
		w.notifierSrv = service.New(w.notifier)
	}

	if !cfg.IgnoreBlockStuckCheck {
		w.updateMonitor = util.NewUpdateMonitor(time.Minute*20, func(v int64, d time.Duration) {
			log.Errorf("%s, block stucks at height %d for %v", workerTag, v, d)
		})
	}

	return w
}

func (w *Worker) Name() string {
	return "deposit"
}

func (w *Worker) Init() error {
	util.Go("tryProcessForceTxs", func() {
		err := w.tryProcessForceTxs()
		if err != nil {
			log.Errorf("%s, try force process txs failed, %v", workerTag, err)
		}
	}, nil)

	if w.notifierSrv != nil {
		go w.notifierSrv.Start()
	}
	return nil
}

func (w *Worker) tryProcessForceTxs() error {
	if len(w.cfg.ForceTxs) == 0 {
		return nil
	}

	log.Warnf("%s, try force process txs: %v", workerTag, w.cfg.ForceTxs)

	txs, err := w.rpcClient.GetTxs(w.cfg.ForceTxs)
	if err != nil {
		return fmt.Errorf("rpc get txs failed, %v", err)
	}

	for _, tx := range txs {
		err = w.processTx(tx)
		if err != nil {
			return fmt.Errorf("force process tx %s failed, %v", tx.Hash, err)
		}
	}

	log.Warnf("%s, finish force process txs: %v", workerTag, w.cfg.ForceTxs)
	return nil
}

// fetch currentBlock from online by rpc
func (w *Worker) Work() {
	if w.updateMonitor != nil {
		defer w.updateMonitor.Check()
	}

	var err error
	defer func() {
		if err != nil {
			time.Sleep(time.Second * 5)
		}
	}()

	if w.currentBlock == nil {
		w.currentBlock, err = w.rpcClient.NextBlock(w.rollbackBlock)
		if err != nil {
			log.Errorf("%s, rpc get block failed, %v", workerTag, err)
			return
		}

		if w.currentBlock == nil {
			time.Sleep(time.Second * 3)
			return
		}
	}

	span := monitor.StartDDSpan("import block", nil, "", monitor.SpanTags{
		"height": w.currentBlock.Height,
		"hash":   w.currentBlock.Hash,
		"txNum":  len(w.currentBlock.Txs),
	})
	defer monitor.DeferFinishDDSpan(span, func() (monitor.SpanTags, error) {
		return nil, err
	})()

	for _, tx := range w.currentBlock.Txs {
		err = w.processTx(tx)
		if err != nil {
			log.Errorf("%s, process tx %s failed, %v", workerTag, tx.Hash, err)
			return
		}
	}

	block := models.BlockInfo{
		Height: w.currentBlock.Height,
		Hash:   w.currentBlock.Hash,
		Symbol: w.cfg.Currency,
	}

	// insert new block to db
	err = block.Insert()
	if err != nil {
		log.Errorf("%s, insert block failed, %v", workerTag, err)
		return
	}

	now := time.Now()
	if now.Sub(w.lastLogInsertBlockTime) > time.Second*3 {
		log.Infof("%s, import new block, height: %d, hash: %s, deposit-txs: %d",
			workerTag, w.currentBlock.Height, w.currentBlock.Hash, len(w.currentBlock.Txs))
		w.lastLogInsertBlockTime = now
	}

	if w.notifier != nil {
		w.notifier.tryNotify()
	}

	if w.updateMonitor != nil {
		w.updateMonitor.Update(int64(w.currentBlock.Height))
	}

	w.currentBlock = nil
}

func (w *Worker) processTx(tx *models.Tx) error {
	if tx == nil {
		log.Errorf("%s, processTx, nil point tx", workerTag)
		return nil
	}

	if tx.Amount.LessThanOrEqual(decimal.New(0, 0)) {
		log.Errorf("%s, processTx, hash: %s, invalid amount: %v", workerTag, tx.Hash, tx.Amount)
		return nil
	}

	if !NormalTxHash(tx.Hash) {
		log.Warnf("abnormal tx hash format: %s", TxString(tx))
	}

	tx.Type = models.TxDeposit
	tx.Amount = tx.Amount.Truncate(TxAmountPrecision)
	tx.Extra = TruncateTxTag(tx.Extra)
	tx.SequenceID = GenSequenceID([]byte(tx.Symbol), []byte(tx.Hash), []byte(tx.Address), []byte(tx.Extra), []byte(strconv.Itoa(int(tx.InnerIndex))))

	if models.TxExistedBySeqID(tx.SequenceID) {
		return nil
	}

	has, err := models.HasAddress(tx.Address)
	if err != nil {
		return err
	}

	if !has {
		return fmt.Errorf("can't find tx.Address from db: %s", tx.Address)
	}

	accept := true
	// ignore deposit if less than min amount.

	if min, _ := currency.MinAmount(tx.Symbol); tx.Amount.LessThan(min) {
		log.Warnf("min=====%s", min)
		accept = false
		log.Infof("%s, ignore min-amount deposit tx, %s", workerTag, TxString(tx))
	}

	if accept && w.cfg.IsNeedTag && !ValidTxTag(tx.Extra, tx.Symbol) {
		accept = false
		log.Infof("%s, ignore invalid-tag in deposit tx, %s", workerTag, TxString(tx))
	}

	if accept && !w.rpcClient.ReuseAddress() {
		if _, err := models.GetTxByAddress(tx.Address); err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
		} else {
			accept = false
			log.Infof("%s, ignore address-reused-tx, %s", workerTag, TxString(tx))
		}
	}

	if !accept {
		tx.NotifyStatus = 1
	}

	// insert new deposit_tx
	err = tx.Insert()
	if err != nil {
		return fmt.Errorf("insert tx %s failed, %v", tx.Hash, err)
	}

	if accept {
		log.Infof("%s, report tx, %s", workerTag, TxString(tx))
	}
	return nil
}

func (w *Worker) rollbackBlock(currentBlock *models.BlockInfo) (*models.BlockInfo, error) {
	if currentBlock == nil {
		return currentBlock, fmt.Errorf("rollback block failed, current block is nil")
	}

	if currentBlock.Height == 0 {
		return currentBlock, fmt.Errorf("rollback block failed, current block height is 0")
	}

	block, ok := models.GetBlockInfoByHeight(currentBlock.Height, w.cfg.Currency, w.cfg.UseBlockTable)
	if !ok {
		return currentBlock, fmt.Errorf("rollback block, get block by height failed, height=%d", currentBlock.Height)
	}

	preBlock, ok := models.GetBlockInfoByHeight(currentBlock.Height-1, w.cfg.Currency, w.cfg.UseBlockTable)
	if !ok {
		return currentBlock, fmt.Errorf("rollback block, get block by height failed, height=%d", currentBlock.Height-1)
	}

	err := block.Delete()
	if err != nil {
		return currentBlock, fmt.Errorf("delete block at height %d failed, %v", block.Height, err)
	}

	log.Warnf("%s, rollback block, height: %d, hash: %s, preHash: %s",
		workerTag, block.Height, block.Hash, preBlock.Hash)
	return preBlock, nil
}

// Destroy
func (w *Worker) Destroy() {
	if w.notifierSrv != nil {
		w.notifierSrv.Stop()
	}
}
