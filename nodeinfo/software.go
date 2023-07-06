package nodeinfo

// Software is metadata about server software in use.
type Software struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Repository string `json:"respository,omitempty"`
	Homepage   string `json:"homepage,omitempty"`
}
