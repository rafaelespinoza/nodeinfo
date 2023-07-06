package nodeinfo

import "encoding/json"

// Protocol describes how a server communicates with other servers.
type Protocol uint

// Protocols enumerated by the NodeInfo spec.
const (
	ActivityPub Protocol = iota + 1
	BuddyCloud
	DFRN
	Diaspora
	Libertree
	OStatus
	PumpIO
	Tent
	XMPP
	Zot
)

var protocolNames = []string{
	"",
	"activitypub",
	"buddycloud",
	"dfrn",
	"diaspora",
	"libertree",
	"ostatus",
	"pumpio",
	"tent",
	"xmpp",
	"zot",
}

func (p Protocol) String() string {
	if int(p) >= len(protocolNames) {
		p = 0
	}

	return protocolNames[p]
}

func (p *Protocol) UnmarshalJSON(in []byte) (err error) {
	var q string
	if err = json.Unmarshal(in, &q); err != nil {
		return
	}

	switch q {
	case protocolNames[1]:
		*p = ActivityPub
	case protocolNames[2]:
		*p = BuddyCloud
	case protocolNames[3]:
		*p = DFRN
	case protocolNames[4]:
		*p = Diaspora
	case protocolNames[5]:
		*p = Libertree
	case protocolNames[6]:
		*p = OStatus
	case protocolNames[7]:
		*p = PumpIO
	case protocolNames[8]:
		*p = Tent
	case protocolNames[9]:
		*p = XMPP
	case protocolNames[10]:
		*p = Zot
	}

	return
}

func (p *Protocol) MarshalJSON() (out []byte, err error) {
	return []byte(`"` + p.String() + `"`), nil
}
