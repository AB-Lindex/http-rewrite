package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	if err := loadConfig(); err != nil {
		slog.Error("Failed loading config", "error", err)
		os.Exit(1)
	}

	var nReg int
	for _, api := range cfg.APIs {
		err := api.Handle()
		if err != nil {
			slog.Error("Failed to register API", "path", api.Input.Path, "error", err)
		} else {
			nReg++
		}
	}
	if nReg == 0 {
		slog.Error("No APIs registered")
		os.Exit(1)
	}

	// remote, err := url.Parse("http://google.com")
	// if err != nil {
	// 	panic(err)
	// }

	// handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	// 	return func(w http.ResponseWriter, r *http.Request) {
	// 		log.Println(r.URL)
	// 		r.Host = remote.Host
	// 		w.Header().Set("X-Ben", "Rad")
	// 		p.ServeHTTP(w, r)
	// 	}
	// }

	slog.Info("Starting server", "port", cfg.ListenPort)

	// proxy := httputil.NewSingleHostReverseProxy(remote)
	// http.HandleFunc("/", handler(proxy))
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.ListenPort), nil)
	if err != nil {
		panic(err)
	}
}
