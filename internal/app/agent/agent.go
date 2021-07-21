package agent

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"registry-cleaner-agent/internal/pkg/fs_analyzer"
	"registry-cleaner-agent/internal/pkg/garbage_collector"
	"registry-cleaner-agent/internal/pkg/registry_api"
	"registry-cleaner-agent/internal/pkg/status"
	"sync"
	"syscall"
	"time"
)

type Agent struct {
	config *Config
	router *mux.Router
	server *http.Server
	gc     *garbage_collector.GCHandler
	wg     *sync.WaitGroup
}

const (
	ShutdownTimeout = 5 * time.Second
)

func New(config *Config) *Agent {
	return &Agent{
		config: config,
		router: mux.NewRouter(),
		wg:     &sync.WaitGroup{},
	}
}

func (a *Agent) Run() {

	err := a.configureRouter()
	if err != nil {
		log.Fatalf("Unable to configure router: %v", err)
	}

	a.configureServer()
	go func() {
		if err = a.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()
	a.waitForShutdownSignal()
	a.shutdown()
}

func (a *Agent) configureServer() {
	ctx, cancel := context.WithCancel(context.Background())
	c := cors.New(cors.Options{
		AllowedOrigins:   a.config.CorsAllowedOrigins,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		AllowedHeaders:   a.config.CorsAllowedHeaders,
		ExposedHeaders:   a.config.CorsExposedHeaders,
	})
	corsHandler := c.Handler(a.router)
	a.server = &http.Server{
		Addr:        a.config.BindAddr,
		Handler:     handlers.RecoveryHandler()(corsHandler),
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}
	a.registerOnShutdown(cancel)
}

func (a *Agent) registerOnShutdown(fn func()) {
	a.wg.Add(1)
	a.server.RegisterOnShutdown(func() {
		go func() {
			defer a.wg.Done()
			fn()
		}()
	})
}

func (a *Agent) waitForShutdownSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(
		ch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	reason := <-ch
	log.Printf("Received [%s] signal\n", reason.String())

	go func() {
		<-ch
		log.Fatal("Terminating immediately\n")
	}()
}

func (a *Agent) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v\n", err)
	} else {
		log.Printf("Agent stopped\n")
	}
	a.wg.Wait()
	log.Println("OnShutdown cleanup finished")

}

func (a *Agent) initHandlers() (*registry_api.RegistryApiHandler, *garbage_collector.GCHandler, error) {
	stm, err := status.InitStatusManager(a.config.BitCaskStoragePath)
	if err != nil {
		return nil, nil, err
	}
	rah, err := registry_api.InitApiHandler(a.config.ApiUrl, stm)
	if err != nil {
		return nil, nil, err
	}
	gc := garbage_collector.NewGarbageCollector(
		a.config.ContainerName, a.config.RegistryConfig)
	fsa := fs_analyzer.NewFSAnalyzer(a.config.RegistryMountPoint)
	gch, err := garbage_collector.InitGCHandler(gc, stm, fsa)
	if err != nil {
		return nil, nil, err
	}

	return rah, gch, nil
}

func (a *Agent) configureRouter() error {
	registryApiHandler, gch, err := a.initHandlers()
	if err != nil {
		return err
	}
	a.router.Use(func(next http.Handler) http.Handler { return handlers.CombinedLoggingHandler(os.Stdout, next) })
	a.router.HandleFunc("/v2/status", registryApiHandler.StatusHandler)
	a.router.HandleFunc("/v2/{repo}/manifests/{tag}/summary", registryApiHandler.ManifestSummaryHandler)

	a.router.HandleFunc("/v2/garbage", gch.GarbageGetHandler).Methods("GET")
	a.router.HandleFunc("/v2/garbage", gch.GarbageDeleteHandler).Methods("DELETE")

	a.router.PathPrefix("/").HandlerFunc(registryApiHandler.ProxyHandler)
	return nil
}
