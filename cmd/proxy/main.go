package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.uber.org/zap"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port   int    `required:"true"`
	Target string `required:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("proxy", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	url, err := url.Parse(cfg.Target)
	if err != nil {
		logger.Error("error parsing backend url",
			zap.String("url", cfg.Target),
			zap.Error(err),
		)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Director = transformRequest(url)
	proxy.ModifyResponse = transformResponse

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Proxy Request",
			zap.String("url", r.URL.String()),
		)
		proxy.ServeHTTP(w, r)
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%v", cfg.Port), nil); err != nil {
		log.Fatal(err.Error())
	}
}

type RequestTransformerFunc func(r *http.Request)

// type ResponseTransformerFunc func(r *http.Response) error

func transformRequest(url *url.URL) RequestTransformerFunc {
	return func(r *http.Request) {
		r.Host = url.Host
		r.URL.Host = url.Host
		r.URL.Scheme = url.Scheme
	}
}

func transformResponse(r *http.Response) error {
	return nil
}
