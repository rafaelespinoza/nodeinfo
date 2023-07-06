package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/rafaelespinoza/nodeinfo/nodeinfo"
)

func doBatchNodeinfo(ctx context.Context, r io.Reader, w io.Writer, c nodeinfo.Client) {
	ctx, cancel := context.WithTimeout(ctx, theArgs.clientTimeout*10) // the timeout should be longer than the client's Timeout.
	defer cancel()

	hrefs, err := readLinesForBatch(ctx, r)
	if err != nil {
		log.Fatalf("failed to read in hostnames: %v", err)
	}

	for r := range batchNodeinfo(ctx, c, hrefs) {
		writeJSONOrDie(r, w)
	}
}

func batchNodeinfo(ctx context.Context, c nodeinfo.Client, hrefs []string) <-chan batchNodeinfoResult {
	// build a single stream of hostnames which can be read by multiple goroutines.
	inputStream := make(chan string)
	go func() {
		defer close(inputStream)

		for _, href := range hrefs {
			inputStream <- href
		}
	}()

	// fan out with a limited number of goroutines. These goroutines read from a
	// shared stream of hostnames until the channel is closed.
	results := make([]<-chan batchNodeinfoResult, maxConcurrentRequests)
	for i := 0; i < maxConcurrentRequests; i++ {
		results[i] = getNodeinfo(ctx, c, inputStream)
	}

	// fan in all results, this simplifies consumption for the caller.
	return mergeBatchNodeinfoResults(ctx, results...)
}

func getNodeinfo(ctx context.Context, c nodeinfo.Client, hostnames <-chan string) <-chan batchNodeinfoResult {
	out := make(chan batchNodeinfoResult)

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

			res := batchNodeinfoResult{HREF: hostname}
			res.NodeInfo, res.Err = c.GetNodeInfo(ctx, res.HREF)
			out <- res
		}
	}()

	return out
}

type batchNodeinfoResult struct {
	HREF     string            `json:"href"`
	NodeInfo nodeinfo.NodeInfo `json:"nodeinfo"`
	Err      error             `json:"err"`
}

func mergeBatchNodeinfoResults(ctx context.Context, inputs ...<-chan batchNodeinfoResult) <-chan batchNodeinfoResult {
	out := make(chan batchNodeinfoResult)

	var wg sync.WaitGroup
	wg.Add(len(inputs))

	for _, in := range inputs {
		go func(results <-chan batchNodeinfoResult) {
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
