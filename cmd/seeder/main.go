package main

import (
	"context"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type config struct {
	Target     string        `envconfig:"TARGET"`
	Interval   time.Duration `envconfig:"INTERVAL"`
	Extensions string        `envconfig:"EXTENSIONS"`
}

func main() {
	var env config
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env config: %v", err)
	}

	client, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("Failed to create cloudevents client: %v", err)
	}

	ext := map[string]string{}
	kvs := strings.Split(env.Extensions, ";")
	for _, kv := range kvs {
		p := strings.Split(kv, ":")
		if len(p) == 2 {
			ext[p[0]] = p[1]
		}
	}

	for {
		e := event.New()
		e.SetID(uuid.New().String())
		e.SetSource("yolocs.ce-test-actor.seeder")
		e.SetType("seed")
		e.SetSubject("tick")
		e.SetTime(time.Now())
		e.SetData("text/plain", "ticking")

		for k, v := range ext {
			e.SetExtension(k, v)
		}

		ret := client.Send(cecontext.WithTarget(context.Background(), env.Target), e)
		if protocol.IsACK(ret) {
			log.Infof("Successfully seeded event (id=%s) to target %q", e.ID(), env.Target)
		} else {
			log.Errorf("Failed to seed event (id=%s) to target %q: %v", e.ID(), env.Target, ret)
		}

		log.Infof("Sleeping...")
		time.Sleep(env.Interval)
	}
}
