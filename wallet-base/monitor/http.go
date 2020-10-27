package monitor

import (
	"fmt"
	"net/http"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

// Info defines simple api response with 200.
func Info(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{}")
}

// ListenAndServe starts a http server for keep alive.
func ListenAndServe(addr string) {
	http.HandleFunc("/", Info)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Errorf("ListenAndServe: %v", err)
	}
}
