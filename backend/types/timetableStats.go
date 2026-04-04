package types

type TimetableFileStats struct {
	LastModified  string `json:"lastModified"`
	ContentLength string `json:"contentLength"`
	Path          string `json:"path"`
}
