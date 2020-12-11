package bsyncd

import (
	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/cmd"
	bsync "upex-wallet/wallet-deposit/syncer/bitcoin"
	"upex-wallet/wallet-deposit/syncer/bitcoin/gbtc"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

func init() {
	cmd.Register("btc", cmd.NewRunType(0, run))
	cmd.Register("ltc", cmd.NewRunType(0, run))
}

func run(cfg *config.Config, restartTimes int) {

	lastBlock := models.GetLastBlockInfo(cfg.Currency, cfg.UseBlockTable)
	log.Infof("init current block, last block, height: %d, hash: %s",
		lastBlock.Height, lastBlock.Hash)
	bitcoinRPC := gbtc.NewClient(cfg.RPCURL)

	log.Infof("init current block, config start height: %d", cfg.StartHeight)
	if cfg.StartHeight > 0 && restartTimes == 0 {
		lastBlock.Height = uint64(cfg.StartHeight)
		lastBlock.Hash = ""
	}

	depositSync := bsync.New(cfg, bitcoinRPC, lastBlock)
	fetcher := bsync.NewFetcher(
		api.NewExAPI(cfg.BrokerURL, cfg.BrokerAccessKey, cfg.BrokerPrivateKey),
		cfg,
		bitcoinRPC,
	)
	depositSync.AddSubscriber(fetcher)
	defer fetcher.Close()

	go fetcher.DepositSchedule()
	depositSync.FetchBlocks()
}
