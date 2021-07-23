package manifest

import (
	"encoding/json"
	"errors"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"io"
	"io/ioutil"
	"net/http"
)

var ErrApiStatusCode = errors.New("docker registry API returned error status")

type Result struct {
	V1Manifest *schema1.Manifest
	V2Manifest *schema2.Manifest
	ApiResp    *http.Response
	Err        error
}

func HeadManifest(manifestUrl string, result chan<- Result) {
	client := &http.Client{}
	req, err := http.NewRequest("HEAD", manifestUrl, nil)
	if err != nil {
		result <- Result{Err: err}
		return
	}
	req.Header.Set("Accept", schema2.MediaTypeManifest)
	resp, err := client.Do(req)
	result <- Result{ApiResp: resp, Err: err}
}

func getManifestData(manifestUrl string, manifestTypeHeader string) (manifestData []byte, apiResp *http.Response, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", manifestUrl, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", manifestTypeHeader)

	apiResp, err = client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(apiResp.Body)

	if apiResp.StatusCode != 200 {
		return nil, apiResp, ErrApiStatusCode
	}

	jsonManifestData, err := ioutil.ReadAll(apiResp.Body)
	return jsonManifestData, apiResp, err
}

func GetV1Manifest(manifestUrl string, result chan<- Result) {
	jsonManifestData, apiResp, err := getManifestData(manifestUrl, schema1.MediaTypeManifest)
	if err != nil {
		result <- Result{ApiResp: apiResp, Err: err}
		return
	}
	manifest := &schema1.Manifest{}
	err = json.Unmarshal(jsonManifestData, manifest)
	result <- Result{V1Manifest: manifest, ApiResp: apiResp, Err: err}
	return
}

func GetV2Manifest(manifestUrl string, result chan<- Result) {
	jsonManifestData, apiResp, err := getManifestData(manifestUrl, schema2.MediaTypeManifest)
	if err != nil {
		result <- Result{ApiResp: apiResp, Err: err}
		return
	}
	manifest := &schema2.Manifest{}
	err = json.Unmarshal(jsonManifestData, manifest)
	result <- Result{V2Manifest: manifest, ApiResp: apiResp, Err: err}
	return
}
