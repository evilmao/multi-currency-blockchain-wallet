package main

import (
	"fmt"
	"time"

	"upex-wallet/wallet-tools/cmd/statuschecker/checker"

	"upex-wallet/wallet-base/cmd"
	"upex-wallet/wallet-base/service"

	"upex-wallet/wallet-base/newbitx/lisc"
)

var (
	configFile    string
	emailPassword string
)

func main() {
	c := cmd.New(
		"statuschecker",
		"check service or blockchain status, and send alert email.",
		"",
		run,
	)
	c.Flags().StringVarP(&configFile, "config", "c", "", "the config file")
	c.Flags().StringVarP(&emailPassword, "emailpassword", "p", "", "the password of sender email")
	c.Execute()
}

func run(c *cmd.Command) error {
	if len(configFile) == 0 {
		return fmt.Errorf("config file can't be empty")
	}

	cfg := lisc.New()
	err := cfg.Load(configFile)
	if err != nil {
		return fmt.Errorf("load config failed, %v", err)
	}

	emailPassword, _ = cfg.String(emailPassword, "emailpassword")
	err = c.Password("emailpassword")
	if err != nil {
		return err
	}

	checker.EmailPassword = emailPassword

	receivers, _ := cfg.Pair.Pair("receivers")
	if receivers != nil {
		for i := 0; i < receivers.ValueCount(); i++ {
			receiver, _ := receivers.String("", i)
			if len(receiver) > 0 {
				checker.Receivers = append(checker.Receivers, receiver)
			}
		}
	}

	alives, _ := cfg.Pair.Pair("alives")
	if alives != nil {
		for i := 0; i < alives.ValueCount(); i++ {
			item, _ := alives.Pair(i)
			if item != nil && item.HasKey() {
				address, _ := item.String("", 0)
				interval, _ := item.Int64(checker.DefaultCheckInterval, 1)

				ck := checker.NewAliveChecker(item.Key(), address)
				go service.NewWithInterval(ck, time.Minute*time.Duration(interval)).Start()
			}
		}
	}

	stucks, _ := cfg.Pair.Pair("stucks")
	if stucks != nil {
		for i := 0; i < stucks.ValueCount(); i++ {
			item, _ := stucks.Pair(i)
			if item != nil && item.HasKey() {
				url, _ := item.String("", 0)
				interval, _ := item.Int64(checker.DefaultCheckInterval, 1)

				ck := checker.NewStuckChecker(item.Key(), url)
				go service.NewWithInterval(ck, time.Minute*time.Duration(interval)).Start()
			}
		}
	}

	upgrades, _ := cfg.Pair.Pair("upgrades")
	if upgrades != nil {
		interval, _ := upgrades.Int64(checker.DefaultCheckInterval, "check-interval")
		urls := map[string]string{}
		for i := 0; i < upgrades.ValueCount(); i++ {
			item, _ := upgrades.Pair(i)
			if item != nil && !item.HasKey() {
				url, _ := item.String("", 0)
				pattern, _ := item.String("", 1)
				urls[url] = pattern
			}
		}

		ck := checker.NewUpgradeChecker(urls)
		go service.NewWithInterval(ck, time.Minute*time.Duration(interval)).Start()
	}

	select {}
}
