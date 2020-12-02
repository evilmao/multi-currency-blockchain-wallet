package cmd

import (
	"strings"
	"time"

	"upex-wallet/wallet-base/service"
	"upex-wallet/wallet-config/deposit/config"
	"upex-wallet/wallet-deposit/cmd"
	"upex-wallet/wallet-deposit/deposit"
	"upex-wallet/wallet-deposit/rpc"
)

func init() {
	cmd.Register("eth", cmd.NewRunType(1, run))
}

func createRPCClient(cfg *config.Config) rpc.RPC {
	if cfg == nil {
		return nil
	}

	creator, ok := rpc.Find(strings.ToUpper(cfg.Currency))
	if !ok {
		return nil
	}
	return creator(cfg)
}

func run(cfg *config.Config, reStartInterval int) {
	rpcClient := createRPCClient(cfg)
	interval := time.Duration(reStartInterval)

	if rpcClient == nil {
		panic("failed to create rpc client")
	}

	depositSrv := service.NewWithInterval(deposit.New(cfg, rpcClient), time.Millisecond*interval)
	defer depositSrv.Stop()
	if err := depositSrv.Start(); err != nil {
		panic(err)
	}
}
