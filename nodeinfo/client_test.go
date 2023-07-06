package nodeinfo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClientDiscoverLinks(t *testing.T) {
	type testcase struct {
		server stubHandler
		expOut []Link
		expErr error
	}

	runTest := func(t *testing.T, test testcase) {
		server := newTestServer(test.server)
		server.StartTLS()
		defer server.Close()

		serverURI, _ := url.Parse(server.URL)

		client := NewClient(0).(*client)
		client.http.Transport = server.Client().Transport

		got, err := client.DiscoverLinks(context.Background(), serverURI.Hostname()+":"+serverURI.Port())
		if test.expErr == nil && err != nil {
			t.Errorf("expected empty error, got %v", err)
		} else if test.expErr != nil && err == nil {
			t.Errorf("expected non-empty error, got %v, exp %v", err, test.expErr)
		} else if test.expErr != nil && err != nil {
			if !errors.Is(err, test.expErr) {
				t.Errorf("expected error (%v) to wrap %v", err, test.expErr)
			}
		}

		if len(got) != len(test.expOut) {
			t.Fatalf("wrong number of links; got %d, exp %d", len(got), len(test.expOut))
		}

		for i, actual := range got {
			expected := test.expOut[i]
			if actual.Rel != expected.Rel {
				t.Errorf("item[%d]; wrong Rel; got %q, exp %q", i, actual.Rel, expected.Rel)
			}
			if actual.HREF != expected.HREF {
				t.Errorf("item[%d]; wrong HREF; got %q, exp %q", i, actual.HREF, expected.HREF)
			}
		}
	}

	t.Run("2xx", func(t *testing.T) {
		runTest(t, testcase{
			server: stubHandler{
				respCode: http.StatusOK,
				respBody: []byte(`{
"links":[
	{"rel":"http://nodeinfo.diaspora.software/ns/schema/2.1","href":"https://example.org/nodeinfo/2.1"},
	{"rel":"http://nodeinfo.diaspora.software/ns/schema/2.0","href":"https://example.org/nodeinfo/2.0"}
]}`),
			},
			expOut: []Link{
				{Rel: "http://nodeinfo.diaspora.software/ns/schema/2.1", HREF: "https://example.org/nodeinfo/2.1"},
				{Rel: "http://nodeinfo.diaspora.software/ns/schema/2.0", HREF: "https://example.org/nodeinfo/2.0"},
			},
			expErr: nil,
		})
	})

	t.Run("4xx", func(t *testing.T) {
		runTest(t, testcase{
			server: stubHandler{
				respCode: http.StatusNotFound,
				respBody: []byte(`"found":false`),
			},
			expErr: errNoProtocolSupport,
		})
	})

	t.Run("5xx", func(t *testing.T) {
		runTest(t, testcase{
			server: stubHandler{
				respCode: http.StatusInternalServerError,
				respBody: []byte(`"error":true`),
			},
			expErr: errRemoteServerError,
		})
	})
}

func TestClientGetNodeInfo(t *testing.T) {
	type testcase struct {
		server stubHandler
		expOut NodeInfo
		expErr error
	}

	runTest := func(t *testing.T, test testcase) {
		server := newTestServer(test.server)
		server.StartTLS()
		defer server.Close()

		serverURI, _ := url.Parse(server.URL)

		client := NewClient(0).(*client)
		client.http.Transport = server.Client().Transport

		reqURI := "https://" + serverURI.Hostname() + ":" + serverURI.Port() + "/nodeinfo/2.0"
		got, err := client.GetNodeInfo(context.Background(), reqURI)
		if test.expErr == nil && err != nil {
			t.Errorf("expected empty error, got %v", err)
		} else if test.expErr != nil && err == nil {
			t.Errorf("expected non-empty error, got %v, exp %v", err, test.expErr)
		} else if test.expErr != nil && err != nil {
			if !errors.Is(err, test.expErr) {
				t.Errorf("expected error (%v) to wrap %v", err, test.expErr)
			}
		}
		t.Logf("%#v", got) // just look it over for now, fields should be non-empty.
	}

	runTest(t, testcase{
		server: stubHandler{
			respCode: http.StatusOK,
			respBody: []byte(`{
"version": "2.0",
"software": {
	"name": "friendica",
	"version": "2023.05-1518"
},
"protocols": [
	"dfrn",
	"activitypub",
	"diaspora"
],
"services": {
	"inbound": [
		"twitter",
		"atom1.0",
		"rss2.0",
		"imap"
	],
	"outbound": [
		"smtp",
		"tumblr",
		"twitter",
		"wordpress",
		"atom1.0"
	]
},
"openRegistrations": true,
"usage": {
	"users": {
		"total": 1926,
		"activeHalfyear": 854,
		"activeMonth": 352
	},
	"localPosts": 79804,
	"localComments": 15922
},
"metadata": {
	"explicitContent": false,
	"nodeName": "social.example.org 🍀"
}
}`),
		},
	})
}

type stubHandler struct {
	timeout  time.Duration
	respCode int
	respBody []byte
}

func newTestServer(s stubHandler) *httptest.Server {
	h := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(s.timeout)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(s.respCode)
		fmt.Fprintf(w, "%s", s.respBody)
	}

	return httptest.NewUnstartedServer(http.HandlerFunc(h))
}
