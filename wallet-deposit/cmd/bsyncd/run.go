package bsyncd

import (
	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/models"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/cmd"
	syncer "upex-wallet/wallet-deposit/syncer/bitcoin"
	"upex-wallet/wallet-deposit/syncer/bitcoin/gbtc"
)

func init() {
	cmd.Register("btc", cmd.NewRunType(0, run))
}

// func main() {
// 	if err := cmd.Execute(run); err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// }

func run(cfg *config.Config, restartTimes int) {

	lastBlock := models.GetLastBlockInfo(cfg.Currency, cfg.UseBlockTable)
	bitcoinRPC := gbtc.NewClient(cfg.RPCURL)

	if cfg.StartHeight > 0 && restartTimes == 0 {
		lastBlock.Height = uint64(cfg.StartHeight)
		lastBlock.Hash = ""
	}

	depositSync := syncer.New(cfg, bitcoinRPC, lastBlock)
	fetcher := syncer.NewFetcher(
		api.NewExAPI(cfg.BrokerURL, cfg.BrokerAccessKey, cfg.BrokerPrivateKey),
		cfg,
		bitcoinRPC,
	)
	depositSync.AddSubscriber(fetcher)
	defer fetcher.Close()

	go fetcher.DepositSchedule()
	depositSync.FetchBlocks()
}
