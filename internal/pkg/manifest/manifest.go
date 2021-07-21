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

func HeadV2Manifest(manifestUrl string) (apiResp *http.Response, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("HEAD", manifestUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", schema2.MediaTypeManifest)
	return client.Do(req)
}

func GetManifestData(manifestUrl string, manifestTypeHeader string) (manifestData []byte, apiResp *http.Response, err error) {
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

func GetV1Manifest(manifestUrl string) (manifest *schema1.Manifest, apiResp *http.Response, err error) {
	jsonManifestData, apiResp, err := GetManifestData(manifestUrl, schema1.MediaTypeManifest)
	if err != nil {
		return nil, apiResp, err
	}
	manifest = &schema1.Manifest{}
	err = json.Unmarshal(jsonManifestData, manifest)
	return manifest, apiResp, err
}

func GetV2Manifest(manifestUrl string) (manifest *schema2.Manifest, apiResp *http.Response, err error) {
	jsonManifestData, apiResp, err := GetManifestData(manifestUrl, schema2.MediaTypeManifest)
	if err != nil {
		return nil, nil, err
	}
	manifest = &schema2.Manifest{}
	err = json.Unmarshal(jsonManifestData, manifest)
	return manifest, apiResp, err
}
