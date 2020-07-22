## Setup

* One namespace `100-100error`
* One broker `testbroker`
* Trigger x 100
* One service per trigger, 100 in total
* One deployment backing the services
  * 10 replicas
  * Each pod has 100% error rate (always return error)
* One seeder that sends one event every second with 100 bytes payload