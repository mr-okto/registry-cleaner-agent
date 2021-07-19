package agent

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"net/http"
	"os"
	"registry-cleaner-agent/internal/pkg/garbage_collector"
	"registry-cleaner-agent/internal/pkg/registry_api"
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
	c := cors.New(cors.Options{
		// INSECURE!
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	})
	return http.ListenAndServe(a.config.BindAddr, handlers.RecoveryHandler()(c.Handler(a.router)))
}

func (a *Agent) initHandlers() (*registry_api.RegistryApiHandler, *garbage_collector.GarbageCollector, error) {
	rah, err := registry_api.Init(a.config.ApiUrl, a.config.BitCaskStoragePath)
	if err != nil {
		return nil, nil, err
	}
	gc := &garbage_collector.GarbageCollector{
		ContainerName:      a.config.ContainerName,
		RegistryConfigPath: a.config.RegistryConfig,
	}
	return rah, gc, nil
}

func (a *Agent) configureRouter() error {
	registryApiHandler, gc, err := a.initHandlers()
	if err != nil {
		return err
	}
	a.router.Use(func(next http.Handler) http.Handler { return handlers.CombinedLoggingHandler(os.Stdout, next) })
	a.router.HandleFunc("/v2/status", registryApiHandler.StatusHandler)
	a.router.HandleFunc("/v2/{repo}/manifests/{tag}/summary", registryApiHandler.MafifestSummaryHandler)

	a.router.HandleFunc("/v2/garbage", gc.GarbageHandler).Methods("GET")
	a.router.HandleFunc("/v2/garbage", gc.GarbageDeleteHandler).Methods("DELETE")

	a.router.PathPrefix("/").HandlerFunc(registryApiHandler.ProxyHandler)
	return nil
}
