package nodeinfo

// Services describes third party sites that a server may connect to via its
// application API.
//
// The specification enumerates values for Inbound and Outbound services, but
// this library does not attempt to represent them as any special identifiers
// because of the fluid nature of social media. Another reason is that some
// values are found in both inbound and outbound services, or are also an
// enumerated protocol; so there would need to be some non-conflicting names.
type Services struct {
	// Inbound is a list of third party sites that a server can retrieve
	// messages from, for combined display with regular traffic.
	Inbound []string `json:"inbound"`

	// Outbound is a list of third party sites that a server can publish
	// messages to on the behalf of a user.
	Outbound []string `json:"outbound"`
}
