# Kuvasz-Streamer

Kuvasz-streamer is an open source change data capture (CDC) project that focuses exclusively on Postgres. It is tightly integrated with Postgres Logical Replication to provide high performance, low latency replication.

## Features

### Lightweight

Kuvasz-streamer is a lightweight service written in Go that has no dependencies and no queuing. Run it as a system service or in a Docker container.

### High-performance

Kuvasz-streamer was benchmarked at 10K tps with less than 1 second latency. It uses the Postgres COPY protocol to perform the initial sync and the logical replication protocol later. It opens multiple connections to the destination database and batches updates into separate transactions.

### Batteries included

Kuvasz-streamer manages publications and replication slots on source databases, adding and deleting configured tables from the publication automatically. It also performs a full sync whenever a new table is added.

### Flexible

Multiple table propagation models are supported: clone, history and append-only.

## Use cases

Kuvasz-streamer can be used for data consolidation, major version upgrades and other cases.

### Microservice database consolidation

In a microservices architecture, each service has its own database. Kuvasz-streamer consolidates all the database of all services into a single data warehouse. The schema in the data warehouse does not have to follow the same one as the original services.

### Multitenant database consolidation

In a sensitive multi-tenant environment, each tenant may be assigned a separate database to ensure that no cross-pollination of data occurs. Kuvasz-streamer can then be used to consolidate all the data in a single table with a tenant identifier to ease reporting.

### Database performance optimization

In a typical microservice architecture, history data is kept to a minimum in order to provide quick query time and low latency to end users. However, historical data is important for AI/ML and reporting. `kuvasz-streamer` implements a no-delete strategy to some tables that does not propagate `DELETE` operations. Example usage includes transaction tables and audit history tables.

### Postgres major version upgrade

Upgrading major versions of Postgres is a time-consuming task that requires substantial downtime. Kuvasz-streamer can be used to synchronize databases between different versions of Postgres and performing a quick switchover.

## Documentation

The documentation is available at https://streamer.kuvasz.io/

## Installation

Check the [Installation Guide](https://streamer.kuvasz.io/installation/) in the documentation.

## Getting started

Detailed instructions are available in the [Getting started](https://streamer.kuvasz.io/getting-started/) section of the documentation

## Discuss

All ideas and discussions are welcome. We use the [GitHub Discussions](https://github.com/kuvasz-io/kuvasz-streamer/discussions) and [Mattermost](https://mattermost.kuvasz.io/signup_user_complete/?id=dxb6abuw3fgj5egbh7cz6gx3yy&md=link&sbr=fa) for that.

### Pull Request Process

Add tests for your changes. 
Ensure the project builds and passes all tests.

