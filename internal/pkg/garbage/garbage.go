package garbage

//easyjson:json
type GarbageBlob struct {
	Size   int    `json:"size"`
	Digest string `json:"digest"`
}

//easyjson:json
type Garbage struct {
	Blobs []GarbageBlob `json:"blobs"`
}

func New() *Garbage {
	return &Garbage{
		Blobs: []GarbageBlob{},
	}
}
