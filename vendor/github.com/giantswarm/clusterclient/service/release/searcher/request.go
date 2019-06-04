package searcher

type Request struct {
	ReleaseVersion string `json:"release_version"`
}

func DefaultRequest() Request {
	return Request{
		ReleaseVersion: "",
	}
}
