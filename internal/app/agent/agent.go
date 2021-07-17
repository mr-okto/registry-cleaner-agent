package agent

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"qoollo-registry-cleaner-agent/internal/pkg/registry_api"
)

type Agent struct {
	config *Config
	router *mux.Router
}

func New(config *Config) *Agent {
	return &Agent{
		config: config,
		router: mux.NewRouter(),
	}
}

func (a *Agent) Start() error {
	err := a.configureRouter()
	if err != nil {
		return err
	}
	return http.ListenAndServe(a.config.BindAddr, handlers.RecoveryHandler()(a.router))
}

func (a *Agent) initHandlers() (*registry_api.RegistryApiHandler, error) {
	rah, err := registry_api.Init(a.config.ApiUrl, a.config.BitCaskStoragePath)
	if err != nil {
		return nil, err
	}
	return rah, nil
}

func (a *Agent) configureRouter() error {
	registryApiHandler, err := a.initHandlers()
	if err != nil {
		return err
	}
	a.router.Use(func(next http.Handler) http.Handler { return handlers.CombinedLoggingHandler(os.Stdout, next) })
	a.router.HandleFunc("/v2/status", registryApiHandler.StatusHandler)
	a.router.PathPrefix("/").HandlerFunc(registryApiHandler.ProxyHandler)
	return nil
}
