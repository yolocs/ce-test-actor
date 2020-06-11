# ce-test-actor

The `actor` accepts events and can be configured to respond differently.

```bash
# (Optional) Set this env to delay the responses.
DELAY=5s
# (Optional) Set this env to only delay for these hosts; use "*" to match all hosts.
DELAY_HOSTS=default-broker.default.svc.cluster.local,other.example.com
# (Optional) Set this env to echo events back for these hosts; use "*" to match all hosts.
ECHO_HOSTS=default-broker.default.svc.cluster.local,other.example.com
# (Optional) Set this env to always respond error for these hosts; use "*" to match all hosts.
ERR_HOSTS=default-broker.default.svc.cluster.local,other.example.com
./actor
```

The `seeder` seeds test events and can be configured to seed differently.

```bash
# (Required) Set this env as the seeding target.
TARGET=default-broker.default.svc.cluster.local
# (Required) Set this env as the interval to seed events.
INTERVAL=1s
# (Default=1) The number of events to seed concurrently each time.
CONCURRENCY=10
# (Optional) Set this env to add additional events extensions.
EXTENSIONS=foo:bar;abc:def
./seeder
```

The `br-gen` helps generate yamls for a test "suite". It generates 1 broker yaml, 1 seeder yaml and 1 triggers yaml. Check out [main.go](./cmd/br-gen/main.go) for flags.

```bash
./br-gen -output=/home/loadtest -ns=loadtest -slow -interval=5m -count=150
```