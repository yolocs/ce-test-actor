package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const brTemplate = `apiVersion: eventing.knative.dev/v1beta1
kind: Broker
metadata:
  name: testbroker
  namespace: {{.namespace}}
  annotations:
    "eventing.knative.dev/broker.class": "googlecloud"
`

// const actorTemplate = `apiVersion: apps/v1
// kind: Deployment
// metadata:
//   name: actor
//   namespace: {{.namespace}}
//   labels:
//     app: actor
// spec:
//   replicas: 3
//   selector:
//     matchLabels:
//       app: actor
//   template:
//     metadata:
//       labels:
//         app: actor
//     spec:
//       containers:
//       - name: actor
//         image: ko://github.com/yolocs/ce-test-actor/cmd/actor
//         ports:
//         - containerPort: 8080
//         env:
// {{.envs}}
// `

const envTemplate = `        - name: {{.envname}}
          value: {{.envvalue}}
`

const trTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: actor-{{.index}}
  namespace: {{.namespace}}
  labels:
    app: actor-{{.index}}
spec:
  replicas: 3
  selector:
    matchLabels:
      app: actor-{{.index}}
  template:
    metadata:
      labels:
        app: actor-{{.index}}
    spec:
      containers:
      - name: actor
        image: ko://github.com/yolocs/ce-test-actor/cmd/actor
        ports:
        - containerPort: 8080
{{.envs}}
---
apiVersion: v1
kind: Service
metadata:
  name: actor-{{.index}}
  namespace: {{.namespace}}
spec:
  selector:
    app: actor-{{.index}}
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
      name: http
---
apiVersion: eventing.knative.dev/v1beta1
kind: Trigger
metadata:
  name: trigger-actor-{{.index}}
  namespace: {{.namespace}}
spec:
  broker: testbroker
  filter:
    attributes:
      type: seed
  subscriber:
    ref:
     apiVersion: v1
     kind: Service
     name: actor-{{.index}}
---

`

const seederTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: seeder
  namespace: {{.namespace}}
  labels:
    app: seeder
spec:
  replicas: 1
  selector:
    matchLabels:
      app: seeder
  template:
    metadata:
      labels:
        app: seeder
    spec:
      containers:
      - name: seeder
        image: ko://github.com/yolocs/ce-test-actor/cmd/seeder
        env:
        - name: TARGET
          value: http://default-brokercell-ingress.cloud-run-events.svc.cluster.local/{{.namespace}}/testbroker
        - name: INTERVAL
          value: {{.interval}}
`

var (
	output       = flag.String("output", "", "Output path")
	ns           = flag.String("ns", "default", "Namesapce")
	count        = flag.Int("count", 100, "The number of triggers to create")
	echo         = flag.Bool("echo", false, "Echo all requests")
	fail         = flag.Bool("fail", false, "Fail all requests")
	slow         = flag.Bool("slow", false, "Timeout all requests")
	seedInternal = flag.String("interval", "5s", "Seed interval")
)

func main() {
	flag.Parse()

	br := strings.ReplaceAll(brTemplate, "{{.namespace}}", *ns)

	// actor := strings.ReplaceAll(actorTemplate, "{{.namespace}}", *ns)
	envs := ""
	if *fail {
		env := strings.ReplaceAll(envTemplate, "{{.envname}}", "ERR_HOSTS")
		env = strings.ReplaceAll(env, "{{.envvalue}}", `"*"`)
		envs += env
	}
	if *slow {
		env1 := strings.ReplaceAll(envTemplate, "{{.envname}}", "DELAY_HOSTS")
		env1 = strings.ReplaceAll(env1, "{{.envvalue}}", `"*"`)
		envs += env1

		env2 := strings.ReplaceAll(envTemplate, "{{.envname}}", "DELAY")
		env2 = strings.ReplaceAll(env2, "{{.envvalue}}", "15m")
		envs += env2
	}
	if *echo {
		env := strings.ReplaceAll(envTemplate, "{{.envname}}", "ECHO_HOSTS")
		env = strings.ReplaceAll(env, "{{.envvalue}}", `"*"`)
		envs += env
	}
	// actor = strings.ReplaceAll(actor, "{{.envs}}", envs)

	triggers := ""
	for i := 0; i < *count; i++ {
		tr := strings.ReplaceAll(trTemplate, "{{.namespace}}", *ns)
		tr = strings.ReplaceAll(tr, "{{.index}}", strconv.Itoa(i))
		if envs != "" {
			envs = "        env:\n" + envs
			tr = strings.ReplaceAll(tr, "{{.envs}}", envs)
		} else {
			tr = strings.ReplaceAll(tr, "{{.envs}}", "")
		}
		triggers += tr
	}

	seeder := strings.ReplaceAll(seederTemplate, "{{.namespace}}", *ns)
	seeder = strings.ReplaceAll(seeder, "{{.interval}}", *seedInternal)

	if err := ioutil.WriteFile(filepath.Join(*output, "broker.yaml"), []byte(br), 0644); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	// if err := ioutil.WriteFile(filepath.Join(*output, "actor.yaml"), []byte(actor), 0644); err != nil {
	// 	log.Println(err)
	// 	os.Exit(1)
	// }
	if err := ioutil.WriteFile(filepath.Join(*output, "triggers.yaml"), []byte(triggers), 0644); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(filepath.Join(*output, "seeder.yaml"), []byte(seeder), 0644); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
