package registry_api

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"qoollo-registry-cleaner-agent/internal/pkg/status"
)

type RegistryApiHandler struct {
	ApiUrl  *url.URL
	Status  *status.Status
	Storage *status.Storage
}

func Init(apiUrl string, storagePath string) (*RegistryApiHandler, error) {
	parsedUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}
	stat := status.New()
	storage := status.Storage{Path: storagePath}
	err = stat.Restore(&storage)
	if err != nil {
		return nil, err
	}
	return &RegistryApiHandler{
		ApiUrl:  parsedUrl,
		Status:  stat,
		Storage: &storage,
	}, nil
}

func (rah *RegistryApiHandler) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(rah.ApiUrl)
	// Update the headers for redirection
	r.URL.Host = rah.ApiUrl.Host
	r.URL.Scheme = rah.ApiUrl.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = r.URL.Host

	proxy.ServeHTTP(w, r)
}

func (rah *RegistryApiHandler) StatusHandler(w http.ResponseWriter, _ *http.Request) {
	healthUrl, err := url.Parse(rah.ApiUrl.String())
	if err != nil {
		// TODO: log errors
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	healthUrl.Path = path.Join(rah.ApiUrl.Path, "/v2/")

	resp, err := http.Get(healthUrl.String())
	rah.Status.IsAlive = err == nil && resp.StatusCode == 200
	res, err := json.Marshal(&rah.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res)
}
