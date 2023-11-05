# orchestrator

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/jspc/pipelines-orchestrator)
[![Go Report Card](https://goreportcard.com/badge/github.com/jspc/pipelines-orchestrator)](https://goreportcard.com/report/github.com/jspc/pipelines-orchestrator)

package orchestrator will, given

1. An input source (such as a database)
2. A way of understanding when a change is made to an input (such as a notify/listen pipe in postgres)
3. A process to run off the back of a change (such as kicking off a docker container)
4. A way of tracking the success of jobs

provide some tooling for orchestrating data pipelines.

This package also exposes a set of interfaces to allow application developers to bring their
inputs and processes

## Types

### type [ContainerImageMissingErr](/container_process.go#L28)

`type ContainerImageMissingErr struct{ ... }`

ContainerImageMissingErr is returned when the ExecutionContext passed to
NewContainerProcess doesn't contain tke key "image"

To fix this, ensure that a container image is set

#### func (ContainerImageMissingErr) [Error](/container_process.go#L36)

`func (e ContainerImageMissingErr) Error() string`

Error implements the error interface and returns a contextual message

This error, while simple and (at least on the face of it) an over-engineered
version of fmt.Errorf("container image missing"), is verbosely implemented
so that callers may use errors.Is(err, orchestrator.ContainerImageMissingErr)
to handle error cases better

### type [ContainerNonZeroExit](/container_process.go#L44)

`type ContainerNonZeroExit int64`

ContainerNonZeroExit is returned when the container exists with anything other
than exit code 0

Container logs should shed light on what went wrong

#### func (ContainerNonZeroExit) [Error](/container_process.go#L47)

`func (e ContainerNonZeroExit) Error() string`

Error returns the error message associated with this error

### type [ContainerProcess](/container_process.go#L52)

`type ContainerProcess struct { ... }`

ContainerProcess allows for processes to be run via a container

#### func [NewContainerProcess](/container_process.go#L62)

`func NewContainerProcess(conf ProcessConfig) (c ContainerProcess, err error)`

NewContainerProcess connects to a container socket, and returns a
ContainerProcess which can be then used to run jobs

#### func (ContainerProcess) [ID](/container_process.go#L81)

`func (c ContainerProcess) ID() string`

ID returns a unique ID for a process manager

#### func (ContainerProcess) [Run](/container_process.go#L86)

`func (c ContainerProcess) Run(ctx context.Context, e Event) (ps ProcessStatus, err error)`

Run takes an Event, and passes it to a container to run

### type [DAG](/dag.go#L54)

`type DAG struct { ... }`

DAG, or 'Directed Acyclic Graph' runs inputs and processes

#### func [New](/dag.go#L63)

`func New() *DAG`

New returns a DAG ready for use

#### func (*DAG) [AddInput](/dag.go#L74)

`func (d *DAG) AddInput(ctx context.Context, i Input) (err error)`

AddInput takes an Input, adds it to the DAG, and runs it
ready for events to flow through

#### func (DAG) [AddLink](/dag.go#L104)

`func (d DAG) AddLink(input Input, process Process) (err error)`

AddLink accepts an Input and a Process, and links them so that when the
input takes a

#### func (DAG) [AddProcess](/dag.go#L95)

`func (d DAG) AddProcess(p Process) error`

AddProcess takes a Process and adds it to the DAG, so that Inputs
can then pass Events to them

### type [Event](/event.go#L8)

`type Event struct { ... }`

Event represents basic metadata that each Input provides

#### func (Event) [JSON](/event.go#L24)

`func (e Event) JSON() (string, error)`

JSON returns the json representation for an event, in a way that our
processes can later work with

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

#### func (InputConfig) [ID](/config.go#L14)

`func (ic InputConfig) ID() string`

ID returns a (hopefully) unique value for this InputConfig

### type [Operation](/operation.go#L20)

`type Operation uint8`

Operation represents one of the basic CRUD operations
on a piece of data and can be used in Inputs to do clever
things around ignoring certain events

#### Constants

```golang
const (
    OperationUnknown Operation = iota
    OperationCreate
    OperationRead
    OperationUpdate
    OperationDelete
)
```

#### func (Operation) [MarshalJSON](/operation.go#L64)

`func (o Operation) MarshalJSON() ([]byte, error)`

MarshalJSON implements the json.Marshaler interface which allows an
Operation to be represented in json (which is really a json string)

#### func (Operation) [MarshalText](/operation.go#L58)

`func (o Operation) MarshalText() (b []byte, err error)`

MarshalText implements the encoding.TextMarshaler interface in order
to get a textual representation of an Operation

#### func (Operation) [String](/operation.go#L70)

`func (o Operation) String() string`

String returns the string representation of an Operation, or
"unknown" for any Operation value it doesn't know about

#### func (*Operation) [UnmarshalJSON](/operation.go#L45)

`func (o *Operation) UnmarshalJSON(b []byte) (err error)`

UnmarshalJSON implements the json.Unmarshaler interface, allowing
for the operation type to be represented in json properly

#### func (*Operation) [UnmarshalText](/operation.go#L25)

`func (o *Operation) UnmarshalText(b []byte) error`

UnmarshalText implements the encoding.TextUnmarshaler interface
allowing for a byte slice containing certain crud operations to be
cast to Operations

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

#### func [NewPostgresInput](/postgres_input.go#L40)

`func NewPostgresInput(ic InputConfig) (p PostgresInput, err error)`

NewPostgresInput accepts an InputConfig and returns a PostgresInput,
which implements the orchestrator.Input interface

The InputConfig.ConnectionString argument can be a DSN, or a postgres
URL

#### func (PostgresInput) [ID](/postgres_input.go#L69)

`func (p PostgresInput) ID() string`

ID returns the ID for this Input

#### func (PostgresInput) [Process](/postgres_input.go#L75)

`func (p PostgresInput) Process(ctx context.Context, c chan Event) (err error)`

Process will configure a database for notification, and then listen to those
notifications

### type [Process](/process.go#L28)

`type Process interface { ... }`

Process is an interface which processes must implement

The inteface is pretty simple: given a

### type [ProcessConfig](/config.go#L20)

`type ProcessConfig struct { ... }`

ProcessConfig contains configuration options for processes, including
an unkeyed map[string]string for arbitrary values

#### func (ProcessConfig) [ID](/config.go#L27)

`func (pc ProcessConfig) ID() string`

ID returns a (hopefully) unique value for this ProcessConfig

### type [ProcessExitStatus](/process.go#L8)

`type ProcessExitStatus uint8`

ProcessExitStatus represents the final status of a Process

#### Constants

```golang
const (
    ProcessUnknown ProcessExitStatus = iota
    ProcessUnstarted
    ProcessSuccess
    ProcessFail
)
```

### type [ProcessInterfaceConversionError](/dag.go#L16)

`type ProcessInterfaceConversionError struct { ... }`

ProcessInterfaceConversionError returns when trying to load a process from our
internal process store returns completely unexpected data

This error represents a huge failure somewhere and should cause a stop-the-world
event

#### func [NewTestProcessInterfaceConversionError](/dag.go#L27)

`func NewTestProcessInterfaceConversionError(input, process string, iface any) ProcessInterfaceConversionError`

NewTestProcessInterfaceConversionError can be used to return a testable error (in tests)

#### func (ProcessInterfaceConversionError) [Error](/dag.go#L22)

`func (e ProcessInterfaceConversionError) Error() string`

Error returns a descriptive error message

### type [ProcessStatus](/process.go#L19)

`type ProcessStatus struct { ... }`

ProcessStatus contains various bits and pieces a process might return,
such as logs and statuscodes and so on

### type [UnknownProcessError](/dag.go#L37)

`type UnknownProcessError struct { ... }`

UnknownProcessError returns when an input tries to trigger a process whch doesn't
exist

#### func [NewTestUnknownProcessError](/dag.go#L46)

`func NewTestUnknownProcessError(input, process string) UnknownProcessError`

#### func (UnknownProcessError) [Error](/dag.go#L42)

`func (e UnknownProcessError) Error() string`

Error returns a descriptive error message

## Sub Packages

* [bin](./bin)

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
