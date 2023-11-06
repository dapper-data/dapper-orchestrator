# orchestrator

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/jspc/pipelines-orchestrator)
[![Go Report Card](https://goreportcard.com/badge/github.com/jspc/pipelines-orchestrator)](https://goreportcard.com/report/github.com/jspc/pipelines-orchestrator)

Package orchestrator provides orchestration and supervision of data pipelines.

These pipelines are made up of inputs and processes; an input is a long running function
which listens to events (such as database triggers, or kafka topics, or webhooks, or anything
really) and a process is a job (such as a docker container, or webhook dispatcher, or kubernetes
job, or anything that runs once) that does something with that event.

By building a system of inputs, processes, and persistence layers it becomes easy to build
sophisticated data pipelines.

For instance, consider the ecommerce analytics pipeline where:

1. A user places an order
2. A backend microservice of some sort places that order into a kafka topic for back-office processing
3. An analytics consumer slurps these messages into a datawarehouse somewhere
4. A write to this datawarehouse is picked up by an input
5. That input triggers a process which runs a series of validations and enrichement processes
6. The process then writes the enriched, validated to a different table in the warehouse
7. That write triggers a different input, which listens for enriched data
8. That input triggers a process which does some kind of final cleansing for gold standard reporting

In effect, this gives us:

```go
[svc] -> (topic() -> [consumer] -> {data warehouse} -> [input -> [encrichment process]] -> {data warehouse} -> [input -> [reporting process]] -> {reporting system}
```

Or, if you like, a way of building a lightweight [medallion architecture]([https://www.databricks.com/glossary/medallion-architecture](https://www.databricks.com/glossary/medallion-architecture))

# When should you use this package?

This package is useful for building customised data pipeline orchestrators, and for building customised components (where off the shelf components, such as Databricks, are fiddly or unable to be customised to the same level).

This package is also useful for running pipelines cheaply, or locally- it requires no outside service (unless you write that into your own service), and doesn't need complicated masters/worker configurations or anything else really

# When should you _not_ use this package?

This package will not give you the same things that off the shelf tools, such as databricks, will give you. There's no easy way to see DAGs, no simple API for updating configuration (unless you write your own).

This package wont do a lot of what you might need; it exists to serve as the engine of a pipeline tool; you must build the rest yourself.

## Types

### type [ContainerImageMissingErr](/container_process.go#L29)

`type ContainerImageMissingErr struct{ ... }`

ContainerImageMissingErr is returned when the ExecutionContext passed to
NewContainerProcess doesn't contain tke key "image"

To fix this, ensure that a container image is set

### type [ContainerNonZeroExit](/container_process.go#L45)

`type ContainerNonZeroExit int64`

ContainerNonZeroExit is returned when the container exists with anything other
than exit code 0

Container logs should shed light on what went wrong

### type [ContainerProcess](/container_process.go#L53)

`type ContainerProcess struct { ... }`

ContainerProcess allows for processes to be run via a container

### type [Event](/event.go#L8)

`type Event struct { ... }`

Event represents basic metadata that each Input provides

### type [Input](/input.go#L12)

`type Input interface { ... }`

Input is a simple interface, and exposes a long running process called Process
which is expected to stream Events.

It is the job of the Orchestrator to understand which channel is assigned to
which input and to route messages accordingly

### type [InputConfig](/config.go#L6)

`type InputConfig struct { ... }`

InputConfig contains the necessary values for coniguring an Input,
such as how to connect to the input source, and the operations the
input supports

### type [Operation](/operation.go#L21)

`type Operation uint8`

Operation represents one of the basic CRUD operations
on a piece of data and can be used in Inputs to do clever
things around ignoring certain events

### type [Orchestrator](/dag.go#L68)

`type Orchestrator struct { ... }`

Orchestrator is the workhorse of this package. It:

```go
1. Supervises inputs
2. Manages the lifecycle of processes, which run on events
3. Syncs events from inputs across multiple processes in a DAG
```

Multiple Orchestrators _can_ be run, like in a cluster, but out of the box
doesn't contain any logic to synchronise inputs and/or processes that wont cluster
natively (such as the postgres sample input)

### type [PostgresInput](/postgres_input.go#L23)

`type PostgresInput struct { ... }`

PostgresInput represents a postgres input source

This source will:

```go
1. Create a function which notifies a channel with a json payload representing an operation
2. Add a trigger to every table in a database to call that function on Creat, Update, and Deletes
3. Listen to the channel created in step 1
```

The operations passed by the database can then be passed to a Process

### type [Process](/process.go#L29)

`type Process interface { ... }`

Process is an interface which processes must implement

The inteface is pretty simple: given a

### type [ProcessConfig](/config.go#L20)

`type ProcessConfig struct { ... }`

ProcessConfig contains configuration options for processes, including
an unkeyed map[string]string for arbitrary values

### type [ProcessExitStatus](/process.go#L8)

`type ProcessExitStatus uint8`

ProcessExitStatus represents the final status of a Process

### type [ProcessInterfaceConversionError](/dag.go#L21)

`type ProcessInterfaceConversionError struct { ... }`

ProcessInterfaceConversionError returns when trying to load a process from our
internal process store returns completely unexpected data

This error represents a huge failure somewhere and should cause a stop-the-world
event

### type [ProcessStatus](/process.go#L20)

`type ProcessStatus struct { ... }`

ProcessStatus contains various bits and pieces a process might return,
such as logs and statuscodes and so on

### type [UnknownProcessError](/dag.go#L42)

`type UnknownProcessError struct { ... }`

UnknownProcessError returns when an input tries to trigger a process whch doesn't
exist

## Sub Packages

* [bin](./bin)

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
