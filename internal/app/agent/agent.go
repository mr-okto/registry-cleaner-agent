package agent

import (
	"errors"
	"github.com/gorilla/mux"
	"net/http"
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
	return http.ListenAndServe(a.config.BindAddr, a.router)
}

func (a *Agent) initHandlers() (*registry_api.RegistryApiHandler, error) {
	rah := registry_api.New(a.config.ApiUrl)
	if rah == nil {
		return nil, errors.New("unable to init registry api handler")
	}
	return rah, nil
}

func (a *Agent) configureRouter() error {
	registryApiHandler, err := a.initHandlers()
	if err != nil {
		return err
	}
	a.router.HandleFunc("/hello", a.handleHello())
	a.router.PathPrefix("/").HandlerFunc(registryApiHandler.Proxy)
	return nil
}

func (a *Agent) handleHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	}
}
