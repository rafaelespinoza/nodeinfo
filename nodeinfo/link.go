package nodeinfo

// A Link contains basic metadata about the server, such as supported versions
// and references to more data. It is part of the initial discovery response
// from a server.
type Link struct {
	Rel  string `json:"rel"`
	HREF string `json:"href"`
}
