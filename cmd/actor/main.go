package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type config struct {
	Delay      time.Duration `envconfig:"DELAY"`
	DelayHosts string        `envconfig:"DELAY_HOSTS"`
	EchoHosts  string        `envconfig:"ECHO_HOSTS"`
	ErrHosts   string        `envconfig:"ERR_HOSTS"`
}

func main() {
	var env config
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env config: %v", err)
	}

	delays := map[string]bool{}
	echos := map[string]bool{}
	errs := map[string]bool{}
	parseHosts(env.DelayHosts, delays)
	parseHosts(env.EchoHosts, echos)
	parseHosts(env.ErrHosts, errs)

	p, err := cloudevents.NewHTTP(cloudevents.WithMiddleware(func(next http.Handler) http.Handler {
		return &reqPrinter{
			next:   next,
			delay:  env.Delay,
			delays: delays,
			errs:   errs,
		}
	}))
	if err != nil {
		log.Fatalf("Failed to create CE HTTP protocol: %v", err)
	}

	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("Failed to create CE client: %v", err)
	}

	log.Error(c.StartReceiver(context.Background(), func(e event.Event) (*event.Event, protocol.Result) {
		log.Infof("Received event: %s", e.String())
		if _, ok := e.Extensions()["actorecho"]; ok {
			return &e, protocol.ResultACK
		}
		return nil, protocol.ResultACK
	}))
}

func parseHosts(hosts string, m map[string]bool) {
	hs := strings.Split(hosts, ",")
	for _, h := range hs {
		m[h] = true
	}
}

type reqPrinter struct {
	next                http.Handler
	delay               time.Duration
	errs, delays, echos map[string]bool
}

func (p *reqPrinter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Infof("Received request %s: %s%s\n", req.Method, req.Host, req.URL.Path)
	log.Infof("Received raw headers: %v", req.Header)

	if _, ok := p.errs[req.Host]; ok {
		http.Error(w, fmt.Sprintf("Injected error for host %q", req.Host), http.StatusBadRequest)
		return
	}

	if _, ok := p.delays[req.Host]; ok {
		log.Infof("Delaying for host %q", req.Host)
		time.Sleep(p.delay)
	}

	if _, ok := p.echos[req.Host]; ok {
		log.Info("Labeling event for echoing back...")
		req.Header.Set("ce-actorecho", "true")
	}

	p.next.ServeHTTP(w, req)
}
