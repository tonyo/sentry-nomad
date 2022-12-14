package main

import (
	"context"
	"math"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/nomad/api"
	log "github.com/sirupsen/logrus"
)

func BeforeSend(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	// Update SDK info
	event.Sdk.Name = "sentry.nomad"
	event.Sdk.Version = "FIXME"

	// Clear modules/packages
	event.Modules = map[string]string{}

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
	log.Debugf("  >> TaskState %+v\n", taskState)

	if taskState.Failed {
		for _, taskEvent := range taskState.Events {
			handleTaskEvent(taskEvent)
		}
	}
}

func handleTaskEvent(taskEvent *api.TaskEvent) {
	log.Debugf("    >> TaskEvent %+v\n", taskEvent)

	// TODO: are event types, hum, types?
	if taskEvent.Type == api.TaskDriverFailure {
		// Report!
		sentryEvent := sentry.Event{Message: taskEvent.DisplayMessage, Level: sentry.LevelError}

		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("eventType", taskEvent.Type)
			sentry.CaptureEvent(&sentryEvent)
		})
	}
}

func handleEvent(event *api.Event) {
	topic := event.Topic
	if topic == api.TopicAllocation {
		alloc, _ := event.Allocation()
		event.Deployment()
		log.Debugf("Allocation: %+v\n", alloc)

		taskStates := alloc.TaskStates

		for _, taskState := range taskStates {
			sentry.WithScope(func(scope *sentry.Scope) {
				// TODO: use SetTags
				scope.SetTag("allocationId", alloc.ID)
				scope.SetTag("allocationName", alloc.Name)
				scope.SetTag("jobId", alloc.JobID)
				scope.SetTag("namespace", alloc.Namespace)
				scope.SetTag("nodeName", alloc.NodeName)
				scope.SetTag("nodeId", alloc.NodeID)
				scope.SetTag("taskGroup", alloc.TaskGroup)
				handleTaskState(taskState)
			})
		}

	} else {
		log.Infof("Skipping event from topic %s\n", topic)
		return
	}
}

func readNomadStream() {
	client, _ := api.NewClient(&api.Config{})
	events := client.EventStream()

	ctx := context.Background()

	// Note: max unsigned (MaxUInt64) triggers a strconv.Atoi "value out of range" error
	const startingIndexMax = uint64(math.MaxInt64)
	eventCh, err := events.Stream(ctx, make(map[api.Topic][]string), startingIndexMax, &api.QueryOptions{})

	if err != nil {
		log.Errorf("Error creating event stream client: %+v err", err)
		os.Exit(1)
	}

	firstEventProcessed := false

	log.Infof("Reading from Nomad event stream...")
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventCh:
			if event.Err != nil {
				log.Errorf("Error from event stream: %+v\n", event)
				// FIXME: add back-off and/or retry
				return
			}
			if event.IsHeartbeat() {
				continue
			}

			for _, e := range event.Events {
				// First event returned from the stream is always an older event, and we want to
				// ignore it.
				if !firstEventProcessed {
					firstEventProcessed = true
					if e.Index >= startingIndexMax {
						log.Errorf("Event index is too big: %d; exiting.\n", e.Index)
						os.Exit(1)
					} else {
						continue
					}
				}
				handleEvent(&e)
			}
		}
	}

}

func main() {
	initSentrySDK()
	readNomadStream()
	log.Info("Done.")
}
