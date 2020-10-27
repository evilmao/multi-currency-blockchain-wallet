package syncer

import (
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-base/monitor"
)

// Subscriber subscribes block/tx message.
type Subscriber interface {
	ImportBlock(block models.BlockInfo) error
	ProcessOrphanBlock(block models.BlockInfo) error
	AddTx(interface{}) error
	AddOrphanTx(interface{}) error
	Close()
}

// DeferTraceImportBlock traces import block span.
func DeferTraceImportBlock(height int, hash string, txNum int, result monitor.DDTraceResult) func() {
	span := monitor.StartDDSpan("import block", nil, "", monitor.SpanTags{
		"height": height,
		"hash":   hash,
		"txNum":  txNum,
	})
	return monitor.DeferFinishDDSpan(span, result)
}
