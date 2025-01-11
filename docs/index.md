---
layout: home
---
Kuvasz-streamer is an open source change data capture (CDC) project that focuses exclusively on Postgres. It is tightly integrated with Postgres Logical Replication to provide high performance, low latency replication.

## Features

### Lightweight

Kuvasz-streamer is a lightweight service written in Go that has zero dependencies and no queuing. Run it as a system service or in a Docker container. It can run in a full declarative mode where the configuration map is stored in a read-only YAML file and no files are written to disk. This mode is suitable for a CI/CD pipeline based configuration and a Kubernetes deployment. An interactive, database-backed mode is supported where the web interface can be used to modify the mapping configuration at runtime.

### High-performance, low latency

Kuvasz-streamer uses the following mechanisms for performance:

- Postgres COPY protocol to perform the initial sync and the logical replication protocol later.
- Multiple parallel connections to the destination database with load sharing.
- Batch updates into periodic transactions.
- Single multi-threaded process with no queuing.
- Rate-limiting on the source connections to avoid source server overload.

Kuvasz-streamer was [benchmarked](https://kuvasz.io/kuvasz-streamer-load-test/) at 10K tps with less than 1 second latency.

### High guarantees

Kuvasz-streamer guarantees

- In-order delivery: changes are applied in the strict order they are received. Although multiple writers are used in parallel, all write to a specific table go to the same writer.
- At-least-once delivery semantics: changes committed on the destination database are relayed back to the source in a status update message. In case of a crash in the streamer or in the destination database, unconfirmed messages are re-applied. Having the same primary keys on the destination and the source guarantees a single application of any update.

### Batteries included

Kuvasz-streamer takes the pain out of managing publications and replications slots:

- It creates missing publications and replications slots on startup
- It adds and removes configured tables from publications automatically
- It performs a full sync whenever a new table is added

It is also fully observable providing [Prometheus metrics]({% link 065-metrics.md %}) and extensive logging.

### Rich streaming modes

Multiple table [streaming modes]({% link 040-streaming-modes.md %}) are supported

- Clone: replicate the source table as-is
- Append-only: replicate the source table but don't delete any records
- History: Keep a full history of all changes with a timestamp

### Full Postgres support

Full PostgreSQL support is guaranteed with an extensive test suite:

- All recent PostgreSQL versions (12 to 17)
- All data types
- Partitions
- Schemas
  - Source tables can be in any database and in any schema
  - Destination tables are in a single database and a single schema

### API and web interface

The service provides an optional API and a web interface to easily manage publications and mapping.

## Use cases

Kuvasz-streamer can be [used](/use-cases/) for data consolidation, major version upgrades and other cases.
