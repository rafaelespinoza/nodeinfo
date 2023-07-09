package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/rafaelespinoza/nodeinfo/nodeinfo"
)

var theArgs struct {
	clientTimeout time.Duration
	batchTimeout  time.Duration
}

func init() {
	const defaultClientTimeout = 5 * time.Second
	flag.DurationVar(&theArgs.clientTimeout, "client-timeout", defaultClientTimeout, "timeout for 1 client request")
	flag.DurationVar(&theArgs.batchTimeout, "batch-timeout", defaultClientTimeout*20, "timeout for entire batch of client requests")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `integration test for nodeinfo client

Usage:
	[flags] <action_name>

	As demonstrated above, any [flags] must be before the <action_name>.
	If invoked like this, <action_name> [flags], then flags are ignored.

	Action name is required. It must be one of
	%q

Examples:
	Discover nodeinfo:

		-client-timeout 5s -batch-timeout 2m batch_discovery

Default flag values:

`, []string{"batch_discovery", "batch_nodeinfo"})
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("expecting 1 positional arg: action")
	}
	action := flag.Args()[0]

	ctx := context.Background()
	out := os.Stdout
	cli := nodeinfo.NewClient(theArgs.clientTimeout)

	if theArgs.batchTimeout <= theArgs.clientTimeout {
		log.Fatalf(
			"-batch-timeout (%s) should be higher than -client-timeout (%s)",
			theArgs.batchTimeout, theArgs.clientTimeout,
		)
	}

	switch action {
	case "batch_discovery":
		doBatchDiscovery(ctx, os.Stdin, out, cli)
	case "batch_nodeinfo":
		doBatchNodeinfo(ctx, os.Stdin, out, cli)
	default:
		log.Fatalf("unexpected action: %q", action)
	}
}

func writeJSONOrDie(in any, w io.Writer) {
	raw, err := json.Marshal(in)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", raw)
}
