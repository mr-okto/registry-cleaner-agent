package registry_api

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"registry-cleaner-agent/internal/pkg/manifest"
	"registry-cleaner-agent/internal/pkg/status"
	"strings"
	"time"
)

type RegistryApiHandler struct {
	ApiUrl        *url.URL
	StatusManager *status.Manager
}

func InitApiHandler(apiUrl string, statusManager *status.Manager) (*RegistryApiHandler, error) {
	parsedUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}
	return &RegistryApiHandler{
		ApiUrl:        parsedUrl,
		StatusManager: statusManager,
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
		log.Printf("[ERROR at RegistryApiHandler.StatusHandler]: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	healthUrl.Path = path.Join(rah.ApiUrl.Path, "/v2/")

	resp, err := http.Get(healthUrl.String())
	rah.StatusManager.Status.IsAlive = err == nil && resp.StatusCode == 200
	res, err := json.Marshal(rah.StatusManager.Status)
	if err != nil {
		log.Printf("[ERROR at RegistryApiHandler.StatusHandler]: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res)
}

func (rah *RegistryApiHandler) ManifestSummaryHeadHandler(w http.ResponseWriter, r *http.Request) {
	manifestUrl := *rah.ApiUrl
	manifestPath := strings.TrimSuffix(r.URL.Path, "/summary")
	manifestUrl.Path = path.Join(manifestUrl.Path, manifestPath)

	resChan := make(chan manifest.Result)
	go manifest.HeadManifest(manifestUrl.String(), resChan)
	res := <-resChan
	close(resChan)

	if res.Err != nil {
		log.Printf("[ERROR at RegistryApiHandler.ManifestSummaryHeadHandler]: %v", res.Err)
		http.Error(w, res.Err.Error(), http.StatusInternalServerError)
		return
	}
	if res.ApiResp.StatusCode != 200 {
		http.Error(w, res.ApiResp.Status, res.ApiResp.StatusCode)
	}
	digest := res.ApiResp.Header.Get("Docker-Content-Digest")
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusOK)
}

func (rah *RegistryApiHandler) ManifestSummaryHandler(w http.ResponseWriter, r *http.Request) {
	manifestUrl := *rah.ApiUrl
	manifestPath := strings.TrimSuffix(r.URL.Path, "/summary")
	manifestUrl.Path = path.Join(manifestUrl.Path, manifestPath)

	v1chan := make(chan manifest.Result)
	go manifest.GetV1Manifest(manifestUrl.String(), v1chan)
	v2chan := make(chan manifest.Result)
	go manifest.GetV2Manifest(manifestUrl.String(), v2chan)
	var results [2]manifest.Result
	results[0] = <-v1chan
	results[1] = <-v2chan
	close(v1chan)
	close(v2chan)
	for _, res := range results {
		if res.Err == manifest.ErrApiStatusCode && res.ApiResp != nil {
			http.Error(w, res.ApiResp.Status, res.ApiResp.StatusCode)
			return
		}
		if res.Err != nil {
			log.Printf("[ERROR at RegistryApiHandler.ManifestSummaryHandler]: %v", res.Err)
			http.Error(w, res.Err.Error(), http.StatusInternalServerError)
			return
		}
	}
	v1Manifest, v2Manifest := results[0].V1Manifest, results[1].V2Manifest
	v2ApiResp := results[1].ApiResp

	imageSize := v2Manifest.Config.Size
	for _, layer := range v2Manifest.Layers {
		imageSize += layer.Size
	}
	creationTime := time.Unix(0, 0).Format(time.RFC3339)
	if len(v1Manifest.History) != 0 {
		value := gjson.Get(v1Manifest.History[0].V1Compatibility, "created")
		creationTime = value.Str
	}
	digest := v2ApiResp.Header.Get("Docker-Content-Digest")
	manifestSummary := manifest.Summary{
		Name:          v1Manifest.Name,
		Tag:           v1Manifest.Tag,
		Architecture:  v1Manifest.Architecture,
		Created:       creationTime,
		Size:          imageSize,
		ContentDigest: digest,
	}

	res, err := json.Marshal(&manifestSummary)
	if err != nil {
		log.Printf("[ERROR at RegistryApiHandler.ManifestSummaryHandler]: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res)
}
