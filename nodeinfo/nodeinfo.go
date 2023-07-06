package nodeinfo

type NodeInfo struct {
	Version           string         `json:"version"` // The schema version.
	Software          Software       `json:"software"`
	Protocols         []Protocol     `json:"protocols"` // The protocols supported on this server.
	Services          Services       `json:"services"`
	OpenRegistrations bool           `json:"openRegistrations"` // Whether the server allows open self-registration.
	Usage             Usage          `json:"usage"`
	Metadata          map[string]any `json:"metadata"` // Free form key value pairs for software specific values. Clients should not rely on any specific key present.
}
