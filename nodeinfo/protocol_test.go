package nodeinfo

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestProtocol(t *testing.T) {
	tests := map[string]Protocol{
		`"activitypub"`: ActivityPub,
		`"buddycloud"`:  BuddyCloud,
		`"dfrn"`:        DFRN,
		`"diaspora"`:    Diaspora,
		`"libertree"`:   Libertree,
		`"ostatus"`:     OStatus,
		`"pumpio"`:      PumpIO,
		`"tent"`:        Tent,
		`"xmpp"`:        XMPP,
		`"zot"`:         Zot,
	}

	t.Run("UnmarshalJSON", func(t *testing.T) {
		for name, protocol := range tests {
			t.Run(name, func(t *testing.T) {
				var got Protocol
				err := json.Unmarshal([]byte(name), &got)

				if err != nil {
					t.Fatal(err)
				}

				if got != protocol {
					t.Errorf("got %q, exp %q", got, protocol)
				}
			})
		}
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		for name, protocol := range tests {
			t.Run(name, func(t *testing.T) {
				out, err := json.Marshal(&protocol)

				if err != nil {
					t.Fatal(err)
				}

				exp := []byte(name)
				same := bytes.Compare(out, exp) == 0
				if !same {
					t.Errorf("got %q, exp %q", out, exp)
				}
			})
		}
	})
}
