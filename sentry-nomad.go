package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/nomad/api"
)

func BeforeSend(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	// Update SDK info
	event.Sdk.Name = "sentry.nomad"
	event.Sdk.Version = "FIXME"

	return event
}

func testSDK() {
	// Using SENTRY_DSN here
	err := sentry.Init(sentry.ClientOptions{
		// Enable printing of SDK debug messages.
		// Useful when getting started or trying to figure something out.
		Debug:            true,
		TracesSampleRate: 0.0,
		BeforeSend:       BeforeSend,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	defer sentry.Flush(2 * time.Second)

	// Test
	sentry.CaptureMessage("It works!")
}

func readNomadStream() {
	client, _ := api.NewClient(&api.Config{})
	events := client.EventStream()

	ctx := context.Background()

	// Note: max unsigned (MaxUInt64) triggers a strconv.Atoi "value out of range" error
	const startingIndex = uint64(math.MaxInt64)
	eventCh, err := events.Stream(ctx, make(map[api.Topic][]string), startingIndex, &api.QueryOptions{})

	if err != nil {
		// s.L.Error("error creating event stream client", "error", err)
		os.Exit(1)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventCh:
			if event.Err != nil {
				// s.L.Warn("error from event stream", "error", err)
				break
			}
			if event.IsHeartbeat() {
				continue
			}

			for _, e := range event.Events {
				fmt.Printf("%+v", e)
			}
		}
	}

}

func main() {
	// testSDK()
	readNomadStream()
	fmt.Println("Done.")
}
