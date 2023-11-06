# Pipeline Orchestrator Example

This directory contains a very simple orchestrator which mimics part of an atmospheric control system at a football stadium; specifically the part of the system which handles precipitation.

Such a system might include a sensor to detect rainfall, readings of which are then used in other places to make decisions (Is it time to cover the pitch? Do we need to do something with under-pitch heating? Do we need to warn players to wear studs, rather than blades?). Such a system might also chuck raw readings from sensors into a Data Warehouse for things like ad-hoc reporting, or deriving Management Information from.

This example orchestrator, then, runs the pipelines for the Data Warehouse side. It consists of two inputs, and two processes:

1. Data is warehoused in the `raw` database (this example doesn't care how that happens, just that it does happen)
2. When data is written to the `raw` database, the `raw_writes` [postgres input](https://pkg.go.dev/github.com/jspc/pipelines-orchestrator#PostgresInput) is triggered
3. This input parses the data from postgres and creates an [Event](https://pkg.go.dev/github.com/jspc/pipelines-orchestrator#Event) which it passes to the orchestrator
4. The orchestrator then finds any process subscribed to this input, which returns a custom [Process](https://pkg.go.dev/github.com/jspc/pipelines-orchestrator#Process) as generated in [writer_process.go](writer_process.go)
5. This process runs a set of validations, ahead of writing to the `cleansed` database (assmuing validations run)
6. Steps 2-4 then run again, only with the `cleansed_writes` [postgres input](https://pkg.go.dev/github.com/jspc/pipelines-orchestrator#PostgresInput) against the `cleansed` database
7. Finally, a neater, split up, enriched set of data is written to the `reporting` database

## Usage

### Prerequisites

You will need postgres running; this tool expects databases available on `postgresql://postgres:postgres@localhost:5432` which can be made available via docker:

```bash
$ docker run --name pipelines-test-db -p 5432:5432 -e POSTGRES_PASSWORD=postgres -d postgres
```

You will also need some initial databases and tables available. Assuming you're running postgres via the above command, this _should_ be as easy as running the included script:

```bash
$ sh ./sql/setup.sh
```

### Building and Running Example

This tool builds and runs exactly as any other go binary:

```bash
$ go build
$ ./example
```

From there, connect to postgres and run the following against the raw database:

```sql
INSERT INTO precipitation (timestamp, location_name, location_latitude, location_longitude, sensor, precipitation) VALUES (now(), 'Anfield', 53.4308435, -2.9633923, 'pitchside-1', 9);
```

Or, via docker:

```bash
cat sql/insert.sql | podman exec -i pipelines-test-db psql -U postgres -d raw
```


### Expected Output

The example binary should output something like:

```
raw_to_cleansed -> valid data <3
cleansed_to_reporting -> created reporting data
```

And your `reporting` database should contain two records:

```bash
$ docker exec -ti pipelines-test-db psql -U postgres
psql (15.3 (Debian 15.3-1.pgdg120+1))
Type "help" for help.

postgres=# \c reporting
You are now connected to database "reporting" as user "postgres".
reporting=# \x
Expanded display is on.
reporting=# select * from sensor;
-[ RECORD 1 ]----------
sensor    | pitchside-1
location  | Anfield
latitude  | 53.4308435
longitude | -2.9633923

reporting=# select * from precipitation;
-[ RECORD 1 ]-------------------------
sensor    | pitchside-1
timestamp | 2023-11-06 14:06:31.670805
value     | 9
```
