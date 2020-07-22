## Setup

* One namespace `100-echo`
* One broker `testbroker`
* Trigger x 100
* One service per trigger, 100 in total
* One deployment backing the services
  * 10 replicas
  * All pods always echo back the same event instantly
* One seeder that sends one event every second with 100 bytes payload