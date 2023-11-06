# orchestrator

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/dapper-data/dapper-orchestrator)
[![Go Report Card](https://goreportcard.com/badge/github.com/dapper-data/dapper-orchestrator)](https://goreportcard.com/report/github.com/dapper-data/dapper-orchestrator)

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

## Sub Packages

* [example](./example)

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
