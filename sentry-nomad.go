package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/nomad/api"
)

func BeforeSend(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	// Update SDK info
	event.Sdk.Name = "sentry.nomad"
	event.Sdk.Version = "FIXME"

	return event
}

func initSentrySDK() {
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
}

func handleTaskState(taskState *api.TaskState) {
	fmt.Printf("  >> TaskState %+v\n", taskState)

	if taskState.Failed {
		for _, taskEvent := range taskState.Events {
			handleTaskEvent(taskEvent)
		}
	}
}

func handleTaskEvent(taskEvent *api.TaskEvent) {
	fmt.Printf("    >> TaskEvent %+v\n", taskEvent)

	// TODO: are event types, hum, types?
	if taskEvent.Type == "Driver Failure" {
		// Report!
		sentryEvent := sentry.Event{Message: taskEvent.DisplayMessage, Level: sentry.LevelError}
		sentry.CaptureEvent(&sentryEvent)
	}
}

func readNomadStream() {
	client, _ := api.NewClient(&api.Config{})
	events := client.EventStream()

	ctx := context.Background()

	// Note: max unsigned (MaxUInt64) triggers a strconv.Atoi "value out of range" error
	const startingIndex = uint64(math.MaxInt64)
	eventCh, err := events.Stream(ctx, make(map[api.Topic][]string), startingIndex, &api.QueryOptions{})

	if err != nil {
		fmt.Printf("Error creating event stream client: %+v err", err)
		os.Exit(1)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventCh:
			if event.Err != nil {
				fmt.Printf("Error from event stream: %+v\n", event)
				// FIXME: add back-off and/or retry
				return
			}
			if event.IsHeartbeat() {
				continue
			}

			for _, e := range event.Events {
				// eventIndex := e.Index
				topic := e.Topic

				if topic == api.TopicAllocation {
					alloc, _ := e.Allocation()
					fmt.Printf("Allocation: %+v\n", alloc)

					taskStates := alloc.TaskStates

					for _, taskState := range taskStates {
						handleTaskState(taskState)
					}

				} else {
					fmt.Printf("-- Skipping event from topic %s\n", topic)
					continue
				}
			}
		}
	}

}

func main() {
	initSentrySDK()
	readNomadStream()
	fmt.Println("Done.")
}
