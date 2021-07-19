package manifest

//easyjson:json
type Summary struct {
	Name          string `json:"name"`
	Tag           string `json:"tag"`
	Architecture  string `json:"architecture"`
	Created       string `json:"created"`
	Size          int64  `json:"size"`
	ContentDigest string `json:"dockerContentDigest"`
}
