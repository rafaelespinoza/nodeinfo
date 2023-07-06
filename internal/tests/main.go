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
	href          string
	clientTimeout time.Duration
}

func init() {
	flag.StringVar(&theArgs.href, "hostname", "example.org", "name of server host")
	flag.StringVar(&theArgs.href, "H", "example.org", "name of server host (short)")
	flag.StringVar(&theArgs.href, "url", "https://example.org/nodeinfo/2.1", "for fetching NodeInfo, the full URL to data")
	flag.StringVar(&theArgs.href, "U", "https://example.org/nodeinfo/2.1", "for fetching NodeInfo, the full URL to data (short)")
	flag.DurationVar(&theArgs.clientTimeout, "client-timeout", 15*time.Second, "timeout for 1 client request")

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

		-hostname 20 discover_one
		discover_one

Default flag values:

`, []string{"discover_one", "get_one", "batch_discovery", "batch_nodeinfo"})
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

	switch action {
	case "discover_one":
		doClientDiscoverOne(ctx, out, cli, theArgs.href)
	case "get_one":
		doClientGetOne(ctx, out, cli, theArgs.href)
	case "batch_discovery":
		doBatchDiscovery(ctx, os.Stdin, out, cli)
	case "batch_nodeinfo":
		doBatchNodeinfo(ctx, os.Stdin, out, cli)
	default:
		log.Fatalf("unexpected action: %q", action)
	}
}

func doClientDiscoverOne(ctx context.Context, w io.Writer, c nodeinfo.Client, hostname string) {
	data, err := c.DiscoverLinks(ctx, hostname)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("got %d results\n", len(data))
	writeJSONOrDie(data, w)
}

func doClientGetOne(ctx context.Context, w io.Writer, c nodeinfo.Client, href string) {
	data, err := c.GetNodeInfo(ctx, href)
	if err != nil {
		log.Fatal(err)
	}

	writeJSONOrDie(data, w)
}

func writeJSONOrDie(in any, w io.Writer) {
	raw, err := json.Marshal(in)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", raw)
}
