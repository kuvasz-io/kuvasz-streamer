---
layout: page
title: Architecture
permalink: /architecture/
nav_order: 90
mermaid: true
---
# Architecture

`kuvasz-streamer` opens a logical replication connection to each of the source databases and a number of connections to the destination database corresponding to the number of workers.

Depending on the running mode, it also reads its mapping configuration either from a static YAML file or from an SQLite database.

![Architecture](/assets/images/architecture.png){: width="50%"}

## Initial synchronization

On startup `kuvasz-streamer` will check the state of the publications and replication slots. If they don't exist, it will create them and initiate a full sync. If the slot exists but some tables have been added, a full sync for these tables only is performed.

![Architecture](/assets/images/initialsync.png)

A separate goroutine is created for each source to handle the initial sync process. Source tables are synchronized sequentially within that source. Parallelizing this would increase the load substantially on the source server but may be something to look at in the future.

## Streaming mode

After it has finished initial synchronization, Kuvasz-streamer enters streaming mode. In this mode, Kuvasz-streamer listens to the Postgres logical replication slot and processes the logical replication records as they arrive.

![Streaming](/assets/images/streaming.png)

For each source database, `kuvasz-streamer` creates a single dedicated Reader goroutine. This goroutine open a Postgres replication connection. It reads replication messages and send updates to the source.

For each worker, `kuvasz-streamer` creates a dedicated worker goroutine that opens a regular connection to the destination. This worker applies changes on the destination database. Having multiple workers allows parallelizing write queries and enhancing performance.

When a replication message (XlogData) is received, `kuvasz-streamer` computes the SQL statement to apply on the destination. Then it selects the worker based on a hash of the source table. It then creates an operation (OP) and sends it to that worker. This mechanism ensures that all changes to a given table are processed in the order they were received.

A worker creates a transaction and uses it as a container for all received messages. After a configurable timeout, usually, 1 second, the transaction is committed and the committed LSN is recorded in a shared map for use by the Reader goroutines.

The Reader goroutines periodically calculate the committed LSN and send a Standby Status Update message to the source. This ensures that these messages are deleted from the replication slot. The Committed LSN is computed to guarantee that all operations from a particular source have been applied on all worker connections.

A simple example:

```mermaid
sequenceDiagram
    participant S1 as Source 1
    participant S2 as Source N
    participant R1 as Reader 1
    participant R2 as Reader N
    participant W1 as Worker 1
    participant W2 as Worker 2
    participant D1 as Destination<br/>Connection 1
    participant D2 as Destination<br/>Connection 2
    W1->>D1: BEGIN
    W2->>D2: BEGIN
    S1->>R1: XlogData T1.insert LSN=1
    R1->>W2: OP: INSERT INTO T1(...)
    W2->>D2: INSERT INTO T1(...)
    Note over W2: S1.WrittenLSN=1
    S2->>R2: XlogData T3.insert LSN=55
    R2->>W1: OP: INSERT INTO T3(...)
    W1->>D1: INSERT INTO T3(...)
    Note over W2: S2.WrittenLSN=55
    S1->>R1: XlogData T2.update LSN=2
    R1->>W1: OP: UPDATE T2 SET ...
    W1->>D2: UPDATE T2 SET ...
    Note over W1: S1.WrittenLSN=2
    S1->>R1: XlogData T1.insert LSN=3
    R1->>W2: OP: INSERT INTO T1(...)
    W2->>D2: INSERT INTO T1(...)
    Note over W2: S1.WritenLSN=3
    W1->>D1: COMMIT
    Note over W1: S1.CommittedLSN=2<br/>S2.CommittedLSN=0
    W2->>D2: COMMIT
    Note over W2: S1.CommittedLSN=3<br/>S2.CommittedLSN=55
    R1->>S1: StandbyStatusUpdate CommittedLSN=3
    R2->>S2: StandbyStatusUpdate CommittedLSN=55
```
