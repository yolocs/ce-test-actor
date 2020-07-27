package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const nsTemplate = `apiVersion: v1
kind: Namespace
metadata:
  name: {{.namespace}}`

const brTemplate = `apiVersion: eventing.knative.dev/v1beta1
kind: Broker
metadata:
  name: testbroker
  namespace: {{.namespace}}
  annotations:
    "eventing.knative.dev/broker.class": "{{.brclass}}"
`

// const actorTemplate = `apiVersion: apps/v1
// kind: Deployment
// metadata:
//   name: actor
//   namespace: {{.namespace}}
//   labels:
//     app: actor
// spec:
//   replicas: {{.replicas}}
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

const trTemplate = `apiVersion: v1
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
apiVersion: apps/v1
kind: Deployment
metadata:
  name: actor-{{.index}}
  namespace: {{.namespace}}
  labels:
    app: actor-{{.index}}
spec:
  replicas: {{.replicas}}
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
        env:
{{.envs}}
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
        - name: SIZE
          value: "{{.size}}"
`

var (
	output        = flag.String("output", "", "Output path")
	ns            = flag.String("ns", "default", "Namesapce")
	count         = flag.Int("count", 100, "The number of triggers to create")
	echo          = flag.Bool("echo", false, "Echo all requests")
	fail          = flag.Int("fail", 0, "Fail requests with the given error rate")
	slow          = flag.String("slow", "", "Delay for all requests")
	seedInternal  = flag.String("interval", "1s", "Seed interval")
	size          = flag.Int64("size", 100, "The size of the event payload")
	brClass       = flag.String("brclass", "googlecloud", "The broker class")
	actorReplicas = flag.Int("actors", 10, "The number of actor replicas")
)

func main() {
	flag.Parse()

	namespace := strings.ReplaceAll(nsTemplate, "{{.namespace}}", *ns)

	br := strings.ReplaceAll(brTemplate, "{{.namespace}}", *ns)
	br = strings.ReplaceAll(br, "{{.brclass}}", *brClass)

	// actor := strings.ReplaceAll(actorTemplate, "{{.namespace}}", *ns)
	// actor = strings.ReplaceAll(actor, "{{.replicas}}", strconv.Itoa(*actorReplicas))
	envs := ""
	if *fail > 0 {
		env1 := strings.ReplaceAll(envTemplate, "{{.envname}}", "ERR_HOSTS")
		env1 = strings.ReplaceAll(env1, "{{.envvalue}}", `"*"`)
		envs += env1

		env2 := strings.ReplaceAll(envTemplate, "{{.envname}}", "ERR_RATE")
		env2 = strings.ReplaceAll(env2, "{{.envvalue}}", fmt.Sprintf(`"%d"`, *fail))
		envs += env2
	}
	if *slow != "" {
		env1 := strings.ReplaceAll(envTemplate, "{{.envname}}", "DELAY_HOSTS")
		env1 = strings.ReplaceAll(env1, "{{.envvalue}}", `"*"`)
		envs += env1

		env2 := strings.ReplaceAll(envTemplate, "{{.envname}}", "DELAY")
		env2 = strings.ReplaceAll(env2, "{{.envvalue}}", *slow)
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
		tr = strings.ReplaceAll(tr, "{{.replicas}}", strconv.Itoa(*actorReplicas))
		tr = strings.ReplaceAll(tr, "{{.envs}}", envs)
		triggers += tr
	}

	seeder := strings.ReplaceAll(seederTemplate, "{{.namespace}}", *ns)
	seeder = strings.ReplaceAll(seeder, "{{.interval}}", *seedInternal)
	seeder = strings.ReplaceAll(seeder, "{{.size}}", fmt.Sprintf("%d", *size))

	if err := ioutil.WriteFile(filepath.Join(*output, "00-namespace.yaml"), []byte(namespace), 0644); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(filepath.Join(*output, "01-broker.yaml"), []byte(br), 0644); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	// if err := ioutil.WriteFile(filepath.Join(*output, "01-actor.yaml"), []byte(actor), 0644); err != nil {
	// 	log.Println(err)
	// 	os.Exit(1)
	// }
	if err := ioutil.WriteFile(filepath.Join(*output, "02-triggers.yaml"), []byte(triggers), 0644); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(filepath.Join(*output, "03-seeder.yaml"), []byte(seeder), 0644); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
