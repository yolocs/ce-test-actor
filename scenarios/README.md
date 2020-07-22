# Test Scenarios

▶️ means there are sample artifacts avaialble for the scenario. Please find in the current directory.

✅ means the scenario could be generated using the tools in this repo.

#### Heath ▶️ ✅

All trigger receivers are responding success instantly.

#### Always fail ▶️ ✅

All trigger receivers are responding error instantly.

#### Fail with certain rate ▶️ ✅

All trigger receivers are responding with certain error rate.

#### Timeout ▶️ ✅

All trigger receivers are not responding before timeout.

#### Events with bigger payload ▶️ ✅

All test events are with a bigger payload size.

#### Echo ▶️ ✅

All trigger receivers are responding with the received events (aka. echo back).

#### Bigger payload + delayed reply ✅

All test events are with a bigger payload size. All trigger receivers are responding with a certain delay.

#### Low error rate + delayed reply ✅

All trigger receivers are responding with a low error rate (2%). Successful responses are delayed before reply.

#### Mix of the workloads above ✅
