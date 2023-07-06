package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/rafaelespinoza/nodeinfo/nodeinfo"
)

func doBatchDiscovery(ctx context.Context, r io.Reader, w io.Writer, c nodeinfo.Client) {
	ctx, cancel := context.WithTimeout(ctx, theArgs.clientTimeout*10) // the timeout should be longer than the client's Timeout.
	defer cancel()

	hostnames, err := readLinesForBatch(ctx, r)
	if err != nil {
		log.Fatalf("failed to read in hostnames: %v", err)
	}

	for r := range batchDiscover(ctx, c, hostnames) {
		writeJSONOrDie(r, w)
	}
}

func readLinesForBatch(ctx context.Context, r io.Reader) (hostnames []string, err error) {
	lineReader := bufio.NewReader(r)
	var line string

	for {
		line, err = lineReader.ReadString('\n')
		if err == io.EOF {
			err = nil
			return
		} else if err != nil {
			return
		}
		hostnames = append(hostnames, strings.TrimSuffix(line, "\n"))
	}
}

const maxConcurrentRequests = 64

func batchDiscover(ctx context.Context, c nodeinfo.Client, hostnames []string) <-chan batchDiscoveryResult {
	// build a single stream of hostnames which can be read by multiple goroutines.
	hostnameStream := make(chan string)
	go func() {
		defer close(hostnameStream)

		for _, hostname := range hostnames {
			hostnameStream <- hostname
		}
	}()

	// fan out with a limited number of goroutines. These goroutines read from a
	// shared stream of hostnames until the channel is closed.
	results := make([]<-chan batchDiscoveryResult, maxConcurrentRequests)
	for i := 0; i < maxConcurrentRequests; i++ {
		results[i] = discoverHosts(ctx, c, hostnameStream)
	}

	// fan in all results, this simplifies consumption for the caller.
	return mergeBatchDiscoveryResults(ctx, results...)
}

func discoverHosts(ctx context.Context, c nodeinfo.Client, hostnames <-chan string) <-chan batchDiscoveryResult {
	out := make(chan batchDiscoveryResult)

	go func() {
		defer close(out)

		for hostname := range hostnames {
			select {
			case <-ctx.Done():
				fmt.Fprintf(os.Stderr, "%v; hostname=%s\n", ctx.Err(), hostname)
				return
			default:
				break
			}

			res := batchDiscoveryResult{Hostname: hostname}
			res.Links, res.Err = c.DiscoverLinks(ctx, res.Hostname)
			out <- res
		}
	}()

	return out
}

type batchDiscoveryResult struct {
	Hostname string          `json:"hostname"`
	Links    []nodeinfo.Link `json:"links"`
	Err      error           `json:"err"`
}

func mergeBatchDiscoveryResults(ctx context.Context, inputs ...<-chan batchDiscoveryResult) <-chan batchDiscoveryResult {
	out := make(chan batchDiscoveryResult)

	var wg sync.WaitGroup
	wg.Add(len(inputs))

	for _, in := range inputs {
		go func(results <-chan batchDiscoveryResult) {
			defer wg.Done()

			for res := range results {
				select {
				case out <- res:
				case <-ctx.Done():
					return
				}
			}
		}(in)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
