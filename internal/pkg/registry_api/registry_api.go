package registry_api

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type RegistryApiHandler struct {
	ApiUrl   *url.URL
	rootPath string
}

func New(apiUrl string) *RegistryApiHandler {
	parsedUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil
	}
	return &RegistryApiHandler{
		ApiUrl:   parsedUrl,
		rootPath: parsedUrl.Path,
	}
}

func (rah *RegistryApiHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(rah.ApiUrl)
	// Update the headers for redirection
	r.URL.Host = rah.ApiUrl.Host
	r.URL.Scheme = rah.ApiUrl.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = r.URL.Host

	proxy.ServeHTTP(w, r)
}
