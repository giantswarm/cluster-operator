package app

type Index struct {
	Entries map[string][]IndexEntry `json:"entries"`
}

type IndexEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
