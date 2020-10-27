package main

const (
	baseURL = "https://rundeck.bitfe.co"

	errorPath   = "/user/error"
	authPath    = "/j_security_check"
	loginPath   = "/user/login"
	listJobPath = "/api/24/project/WalletDeploy/jobs"
)

func deployPath() string {
	return "/api/24/job/" + jobID + "/executions"
}
