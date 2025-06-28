package api

import (
	"net/http"
)

type Api struct {
	config *Config
}
type Config struct {
	adrr string
}

func NewApi(addr string) *Api {
	return &Api{
		config: &Config{
			adrr: addr,
		},
	}
}

func (a *Api) Run(r http.Handler) error {

	server := http.Server{
		Addr:    a.config.adrr,
		Handler: r,
	}

	return server.ListenAndServe()

}
