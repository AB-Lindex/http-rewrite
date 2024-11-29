package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/ninlil/envsubst"
)

type proxyConfig struct {
	ListenPort int         `yaml:"listen"`
	APIs       []*ProxyAPI `yaml:"apis"`
}

type ProxyAPI struct {
	Input proxyInput    `yaml:"input"`
	Proxy *proxyTarget  `yaml:"proxy"`
	Http  *statusTarget `yaml:"http"`
}

type statusTarget struct {
	Status int    `yaml:"status"`
	Body   string `yaml:"body"`
}

type proxyInput struct {
	Method  string   `yaml:"method"`
	Methods []string `yaml:"methods"`
	Path    string   `yaml:"path"`
}

type proxyTarget struct {
	Scheme string      `yaml:"scheme"`
	Host   string      `yaml:"host"`
	Port   int         `yaml:"port"`
	Path   string      `yaml:"path"`
	Query  *ProxyQuery `yaml:"query"`
	remote *httputil.ReverseProxy
}

type ProxyQuery struct {
	Set map[string]string `yaml:"set"`
}

var cfg proxyConfig

func (api *ProxyAPI) Handle(r *chi.Mux) error {

	if api.Input.Method != "" && len(api.Input.Methods) > 0 {
		return fmt.Errorf("only one of method or methods can be specified")
	}

	if api.Proxy == nil && api.Http == nil {
		return fmt.Errorf("either proxy or http must be specified")
	}

	var handler http.HandlerFunc = nil
	if api.Proxy != nil {

		if api.Proxy.Scheme == "" {
			api.Proxy.Scheme = "http"
		}
		if api.Proxy.Host == "" {
			api.Proxy.Host = "localhost"
		}
		if api.Proxy.Port == 0 {
			return fmt.Errorf("port is required")
		}

		remote, err := url.Parse(fmt.Sprintf("%s://%s:%d", api.Proxy.Scheme, api.Proxy.Host, api.Proxy.Port))
		if err != nil {
			panic(err)
		}
		api.Proxy.remote = httputil.NewSingleHostReverseProxy(remote)
		handler = api.ProxyHandler
	}

	if api.Http != nil {
		handler = func(w http.ResponseWriter, r *http.Request) {
			slog.Info(fmt.Sprintf("Status url: %s %s", r.Method, r.URL), "status", api.Http.Status)
			w.WriteHeader(api.Http.Status)
			if api.Http.Body != "" {
				w.Write([]byte(api.Http.Body))
			}
		}
	}

	methods := make(map[string]bool)
	if api.Input.Method != "" {
		methods[api.Input.Method] = true
	}
	for _, m := range api.Input.Methods {
		methods[m] = true
	}
	all := methods["*"] || methods["ALL"]

	if all {
		slog.Info("Registering wildcard-API", "pattern", api.Input.Path)
		r.HandleFunc(api.Input.Path, handler)
	} else {
		for m := range methods {
			// pattern := fmt.Sprintf("%s %s", m, api.Input.Path)
			// slog.Info("Registering API", "pattern", pattern)
			slog.Info("Registering API", "method", m, "pattern", api.Input.Path)
			r.MethodFunc(m, api.Input.Path, handler)
		}
	}
	return nil
}

func (api *ProxyAPI) ProxyHandler(w http.ResponseWriter, r *http.Request) {

	mapper := api.Mapper(r)

	if api.Proxy.Path != "" {
		newpath, err := envsubst.ConvertString(api.Proxy.Path, mapper)
		if err != nil {
			slog.Error("Failed to substitute path", "error", err, "request", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.URL.Path = newpath
	}

	if api.Proxy.Query != nil {
		query := r.URL.Query()

		if len(api.Proxy.Query.Set) > 0 {
			for k, v := range api.Proxy.Query.Set {
				newV, err := envsubst.ConvertString(v, mapper)
				if err != nil {
					slog.Error("Failed to substitute query", "param", k, "error", err, "request", r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				query.Set(k, newV)
			}
		}

		r.URL.RawQuery = query.Encode()
	}

	w2 := &writer{w: w}
	api.Proxy.remote.ServeHTTP(w2, r)
	slog.Info(fmt.Sprintf("Proxy url: %s %s", r.Method, r.URL), "status", w2.Status)
}

func (api *ProxyAPI) Mapper(r *http.Request) func(string) (string, bool) {
	return func(key string) (string, bool) {
		if v := chi.URLParam(r, key); v != "" {
			return v, true
		}
		if v := r.URL.Query().Get(key); v != "" {
			return v, true
		}
		if key == "*" {
			return "", true
		}
		return "", false
	}
}
