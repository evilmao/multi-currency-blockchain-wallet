package main

import (
	"errors"

	"upex-wallet/wallet-base/api"
	"upex-wallet/wallet-base/newbitx/misclib/eosio"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/cmd"
	syncer "upex-wallet/wallet-deposit/syncer/eos"
)

func init() {
	cmd.Register("eos", cmd.NewRunType(0, run))
}

func run(cfg *config.Config, restartTimes int) {
	rpcClient := eosio.New(cfg.RPCURL)
	if rpcClient == nil {
		panic(errors.New("rpc connect error"))
	}

	depositSync := syncer.New(cfg, rpcClient)
	fetcher := syncer.NewFetcher(
		api.NewExAPI(cfg.BrokerURL, cfg.BrokerAccessKey, cfg.BrokerPrivateKey),
		cfg,
		rpcClient,
	)

	depositSync.AddSubscriber(fetcher)
	defer fetcher.Close()

	go fetcher.DepositSchedule()
	depositSync.FetchBlocks()
}
