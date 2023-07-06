package nodeinfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client interface {
	DiscoverLinks(ctx context.Context, hostname string) (out []Link, err error)
	GetNodeInfo(ctx context.Context, href string) (out NodeInfo, err error)
}

func NewClient(timeout time.Duration) Client {
	var h http.Client
	h.Timeout = timeout
	return &client{http: &h}
}

type client struct {
	http *http.Client
}

// discoveryPath is the path on the remote server that would contain links to
// NodeInfo data. This path is required by the NodeInfo spec.
const discoveryPath = ".well-known/nodeinfo"

var (
	errNoProtocolSupport = errors.New("host does not support NodeInfo protocol")
	errRemoteServerError = errors.New("remote server error")
)

func (c *client) DiscoverLinks(ctx context.Context, hostname string) (out []Link, err error) {
	reqURI, err := url.JoinPath("https://", hostname, discoveryPath)
	if err != nil {
		return
	}
	resp, err := c.get(ctx, reqURI)
	if err != nil {
		return
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		// The NodeInfo protocol says:
		//
		// > A client should abandon the discovery on a HTTP response status
		// > code of 404 or 400 and may mark the host as not supporting the
		// > NodeInfo protocol.
		err = fmt.Errorf("%w, status_code=%d", errNoProtocolSupport, resp.StatusCode)
		return
	} else if resp.StatusCode >= 500 {
		// The NodeInfo protocol says:
		//
		// > A client should retry discovery on server errors as indicated by
		// > the HTTP response status code 500.
		err = fmt.Errorf("%w, client should retry later, status_code=%d", errRemoteServerError, resp.StatusCode)
		return
	}
	// TODO: implement error handling for HTTPS connection errors.
	// If HTTPS has connection errors, retry with HTTP
	// If the response is 4xx, then abandon. Mark host as not supporting the
	// NodeInfo protocol.

	var body struct {
		Links []Link `json:"links"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return
	}
	out = body.Links
	return
}

func (c *client) GetNodeInfo(ctx context.Context, href string) (out NodeInfo, err error) {
	reqURI, err := url.ParseRequestURI(href)
	if err != nil {
		return
	}

	resp, err := c.get(ctx, reqURI.String())
	if err != nil {
		return
	}

	defer func() { _ = resp.Body.Close() }()

	// TODO: Handle 4xx, 5xx response codes. Maybe differently than how DiscoverLinks does.

	err = json.NewDecoder(resp.Body).Decode(&out)
	return
}

func (c *client) get(ctx context.Context, reqURI string) (resp *http.Response, err error) {
	fmt.Fprintf(os.Stderr, "nodeinfo: req_uri=%s\n", reqURI) // TODO: make this an option

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURI, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/json")

	resp, err = c.http.Do(req)
	return
}
