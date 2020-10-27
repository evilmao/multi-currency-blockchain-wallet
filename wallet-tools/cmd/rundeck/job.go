package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"upex-wallet/wallet-base/util"

	"github.com/buger/jsonparser"
)

type deployJob struct {
	client  *http.Client
	version string
	job     string
}

func newJob(version, job string) *deployJob {
	return &deployJob{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Jar: newCookieJar(),
		},
		version: version,
		job:     job,
	}
}

func (j *deployJob) run() error {
	err := util.Try(3, func(int) error {
		return j.auth()
	})
	if err != nil {
		return err
	}
	return j.runJob()
}

func (j *deployJob) auth() error {
	values := url.Values{
		"j_username": {username},
		"j_password": {password},
	}

	resp, err := j.client.PostForm(baseURL+authPath, values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if len(location) == 0 || strings.Contains(location, loginPath) || strings.Contains(location, errorPath) {
		return fmt.Errorf("auth failed")
	}

	return nil
}

func (j *deployJob) runJob() error {
	fmt.Printf("start deploy %s %s (%s) at %s:\n", j.job, j.version, jobID, time.Now().Format("2006-01-02 15:04:05"))

	args := fmt.Sprintf("{\"argString\":\"-Version %s -Service %s\"}", j.version, j.job)
	req, err := http.NewRequest("POST", baseURL+deployPath(), strings.NewReader(args))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	failed, _ := jsonparser.GetBoolean(data, "error")
	if failed {
		return fmt.Errorf(string(data))
	}

	fmt.Println(string(data))
	fmt.Println()

	return nil
}

func (j *deployJob) listJobs() error {
	req, err := http.NewRequest("GET", baseURL+listJobPath, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}
