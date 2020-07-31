package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/netutil"
)

type config struct {
	Delay      time.Duration `envconfig:"DELAY"`
	DelayHosts string        `envconfig:"DELAY_HOSTS"`
	EchoHosts  string        `envconfig:"ECHO_HOSTS"`
	ErrHosts   string        `envconfig:"ERR_HOSTS"`
	ErrRate    int           `envconfig:"ERR_RATE"`
	MaxConn    int           `envconfig:"MAX_CONN"`
}

func main() {
	// Create a large heap allocation of 10 GiB
	_ = make([]byte, 10<<30)

	var env config
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env config: %v", err)
	}

	delays := parseHosts(env.DelayHosts)
	echos := parseHosts(env.EchoHosts)
	errs := parseHosts(env.ErrHosts)

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to create net listener: %v", err)
	}
	defer l.Close()
	if env.MaxConn > 0 {
		l = netutil.LimitListener(l, env.MaxConn)
	}

	p, err := cloudevents.NewHTTP(
		cloudevents.WithListener(l),
		cloudevents.WithMiddleware(func(next http.Handler) http.Handler {
			return &reqPrinter{
				next:    next,
				delay:   env.Delay,
				delays:  delays,
				errs:    errs,
				echos:   echos,
				errRate: env.ErrRate,
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
		log.Infof("Latency: %v seconds", time.Now().Sub(e.Time()).Seconds())
		if _, ok := e.Extensions()["actorecho"]; ok {
			return &e, protocol.ResultACK
		}
		return nil, protocol.ResultACK
	}))
}

func parseHosts(hosts string) *matchHosts {
	if hosts == "*" {
		return &matchHosts{matchAll: true}
	}

	m := make(map[string]bool)
	hs := strings.Split(hosts, ",")
	for _, h := range hs {
		m[h] = true
	}

	return &matchHosts{whitelist: m}
}

type reqPrinter struct {
	next                http.Handler
	delay               time.Duration
	errs, delays, echos *matchHosts
	errRate             int
	concurr             int64
}

func (p *reqPrinter) diceErr() bool {
	if p.errRate <= 0 && p.errRate > 100 {
		return true
	}
	r := rand.Int31n(100)
	return int(r) < p.errRate
}

func (p *reqPrinter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	atomic.AddInt64(&p.concurr, 1)
	startTime := time.Now()
	defer func() {
		log.Infof("Done request in %vs event id %s; handling %d requests concurrently", time.Now().Sub(startTime).Seconds(), req.Header.Get("Ce-Id"), atomic.LoadInt64(&p.concurr))
		atomic.AddInt64(&p.concurr, -1)
	}()
	log.Infof("New request event id %s; handling %d requests concurrently", req.Header.Get("Ce-Id"), atomic.LoadInt64(&p.concurr))
	log.Infof("Received request %s: %s%s\n", req.Method, req.Host, req.URL.Path)
	log.Infof("Received raw headers: %v", req.Header)

	if p.errs.include(req.Host) {
		if p.diceErr() {
			w.Header().Set("content-type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Injected error for host %q", req.Host)))
			return
		}
	}

	if p.delays.include(req.Host) {
		log.Infof("Delaying for host %q", req.Host)
		// The actual sleep time is NOT always precise, especially under load.
		time.Sleep(p.delay)
	}

	if p.echos.include(req.Host) {
		log.Info("Labeling event for echoing back...")
		req.Header.Set("ce-actorecho", "true")
	}

	p.next.ServeHTTP(w, req)
}

type matchHosts struct {
	matchAll  bool
	whitelist map[string]bool
}

func (m *matchHosts) include(h string) bool {
	if m.matchAll {
		return true
	}
	_, ok := m.whitelist[h]
	return ok
}
