package main

import (
	"fmt"
	"net/http"
	"time"

	"upex-wallet/wallet-base/cmd"

	"upex-wallet/wallet-base/newbitx/lisc"
)

var (
	configFile string
	username   string
	password   string
	version    string
	versions   []string
	services   []string
	jobID      string

	httpClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: newCookieJar(),
	}
)

func main() {
	c := cmd.New(
		"rundeck",
		"deploy rundeck jobs for deposit.",
		"./rundeck -u username -p password -i job-id -v v0.3.9.3 -s wallet-deposit-btc,wallet-deposit-eth",
		run)
	c.Flags().StringVarP(&configFile, "config", "c", "", "the config file")
	c.Flags().StringVarP(&username, "username", "u", "", "the login username")
	c.Flags().StringVarP(&password, "password", "p", "", "the login password, for inputting in security mode, provide an empty string")
	c.Flags().StringVarP(&version, "version", "v", "", "the version to deploy")
	c.Flags().StringSliceVarP(&services, "services", "s", nil, "the services to deploy")
	c.Flags().StringVarP(&jobID, "job", "i", "", "deploy job id")
	c.Execute()
}

func run(c *cmd.Command) error {
	err := initConfig()
	if err != nil {
		return err
	}

	err = c.Password("password")
	if err != nil {
		return err
	}

	err = checkEmpty()
	if err != nil {
		return err
	}

	for i, service := range services {
		job := newJob(versions[i], service)
		err := job.run()
		if err != nil {
			fmt.Printf("deploy %s failed, %v\n\n", service, err)
			continue
		}

		if i < len(services)-1 {
			time.Sleep(time.Second * 75)
		}
	}

	return nil
}

func initConfig() error {
	if len(configFile) == 0 {
		return nil
	}

	cfg := lisc.New()
	err := cfg.Load(configFile)
	if err != nil {
		return err
	}

	username, _ = cfg.String(username, "username")
	password, _ = cfg.String(password, "password")
	version, _ = cfg.String(version, "version")

	for i := 0; i < len(services); i++ {
		versions = append(versions, version)
	}

	pair, _ := cfg.Pair.Pair("services")
	if pair != nil {
		for i := 0; i < pair.ValueCount(); i++ {
			item, _ := pair.Value(i)
			var serv string
			var vers = version
			if item.Type() == lisc.StringType {
				serv, _ = pair.String("", i)
			} else if item.Type() == lisc.PairType {
				serv = item.(*lisc.Pair).Key()
				vers, _ = pair.String(version, i)
				if len(vers) == 0 {
					vers = version
				}
			}

			if len(serv) > 0 {
				services = append(services, serv)
				versions = append(versions, vers)
			}
		}
	}

	jobID, _ = cfg.String("", "job")
	return nil
}

func checkEmpty() error {
	fields := map[string]string{
		"username": username,
		"password": password,
		"job":      jobID,
	}

	for k, v := range fields {
		if len(v) == 0 {
			return fmt.Errorf("%s can't be empty", k)
		}
	}

	if len(services) == 0 {
		return fmt.Errorf("services can't be empty")
	}

	return nil
}
