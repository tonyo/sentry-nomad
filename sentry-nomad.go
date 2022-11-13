package main

import (
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
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

func main() {
	testSDK()
	fmt.Println("Done.")
}
