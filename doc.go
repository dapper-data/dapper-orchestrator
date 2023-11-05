/*
package orchestrator will, given

1. An input source (such as a database)
2. A way of understanding when a change is made to an input (such as a notify/listen pipe in postgres)
3. A process to run off the back of a change (such as kicking off a docker container)
4. A way of tracking the success of jobs

provide some tooling for orchestrating data pipelines.

This package also exposes a set of interfaces to allow application developers to bring their
inputs and processes
*/
package orchestrator
