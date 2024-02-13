---
# Feel free to add content and custom Front Matter to this file.
# To modify the layout, see https://jekyllrb.com/docs/themes/#overriding-theme-defaults

layout: home
---
# Kuvasz-Streamer

Kuvasz-streamer is an open source change data capture (CDC) project that focuses exclusively on Postgres. It is tightly integrated with Postgres Logical Replication to provide high performance, low latency replication.

## Use cases

Kuvasz-streamer can be used for data consolidation, major verison upgrades and other cases.

### Microservice database consolidation

In a microservices architecture, each service has its own database. Kuvasz-streamer consolidates all the database of all services into a single data warehouse. The schema in the data warehouse does not have to follow the same one as the original services.

### Multitenant database consolidation

In a sensitive multi-tenant environment, each tenant may be assigned a separate database to ensure that no cross-pollination of data occurs. Kuvasz-streamer can then be used to consolidate all the data in a single table with a tenant identifier to ease reporting.

### Database performance optimization

In a typical microservice architecture, history data is kept to a minimum in order to provide quick query time and low latency to end users. However, historical data is important for AI/ML and reporting. `kuvasz-streamer` implements a no-delete strategy to some tables that dows not propagate `DELETE` operations. Example usage includes transaction tables and audit history tables.

### Postgres major version upgrade

Upgrading major versions of Postgres is a time-consuming task that requires substantial downtime. Kuvasz-streamer can be used to synchronize databases between different versions of Postgres and performing a quick switchover.

## Features

### Lightweight

Kuvasz-streamer is a lightweight service written in Go that has no dependencies.

### High-performance

Kuvasz-streamer uses the Postgres COPY protocol to perform the initial sync and the logical replication protocol later.

### Flexible

Multiple table propagation models are supported as described below.

## Database operation propagation

### type = `clone`
These are table that need to be identical between source and destination and where historical data is not important. Example: product types, colors.

- INSERT
  ```sql
  INSERT INTO destination(sid, ...) 
  VALUES (SID, ...)
  ```
  If key already exists, log error
- UPDATE
  ```sql
  UPDATE destination
  SET ...=... 
  WHERE sid=SID and PK=...
  ```
  If key does not exist: insert row and log error.
- DELETE
  ```sql
  DELETE FROM destination
  WHERE sid=SID AND PK=...
  ```
  If key does not exist: log error.

### type = `append`
In high-performance systems, it is important to keep a small number of historical events in the live service table and keep a much larger history in a data warfehouse. Examples: audit events, notifications, transactions. 

These tables behave the same as `clone` tables with the exception that `DELETE` and `TRUNCATE` are ignored.

### type = `history`
History tables implement Slowly Changing Dimensions (SCD) type 2. They are useful to keep a complete history of all changes. Examples include changes in the salary field of an employee. History tables should be used carefully as they generate a lot of rows and the destination table may grow out of control.

More information can be found in this [Wikipedia Article](https://en.wikipedia.org/wiki/Slowly_changing_dimension)

- INSERT
  ```sql
  INSERT INTO destination(sid, ..., kvsz_start, kvsz_end, kvsz_deleted) 
  VALUES(SID, ...,  '1900-01-01', '9999-01-01', false)
  ```
  - If key already exists, log error
- UPDATE
  ```sql
  UPDATE destination 
  SET kvsz_end=now()
  WHERE sid=SID AND kvsz_end='9999-01-01' AND PK=...
  ```
  ```sql
  INSERT INTO destination(sid, ..., kvsz_start, kvsz_end, kvsz_deleted)
  VALUES(SID, ..., now(), '9999-01-01', false)
  ```
  If key does not exist, just insert the row and log error.
- DELETE
  ```sql
  UPDATE destination
  SET kvsz_end=now(), kvsz_deleted=true
  WHERE sid=SID AND kvsz_end='9999-01-01' AND PK=...
  ```
  If key does not exist, log error.
- SELECT latest values
  ```sql
  SELECT *
  FROM destination
  WHERE sid=SID and id=ID and kvsz_end='9999-01-01'
- SELECT historical values
  ```sql
  SELECT *
  FROM destination
  WHERE sid=SID and id=ID and '2023-01-28' between kvsz_start and kvsz_end
  ```

## History table example

### Add record 2020-01-01, salary=1000

|sid|id|first_name|last_name|salary|kvsz_start|kvsz_end|kvsz_deleted|
|---|--|----------|---------|------|----------|--------|------------|
|1|1|John|Doe|1000|1900-01-01|9999-01-01|false


### Update record on 2023-01-01, salary=1200

|sid|id|first_name|last_name|salary|kvsz_start|kvsz_end|kvsz_deleted|
|---|--|----------|---------|------|----------|--------|------------|
|1|1|John|Doe|1000|1900-01-01|2024-01-01|false
|1|1|John|Doe|1000|2024-01-01|9999-01-01|false

### Update record on 2024-01-01, salary=2000

|sid|id|first_name|last_name|salary|kvsz_start|kvsz_end|kvsz_deleted|
|---|--|----------|---------|------|----------|--------|------------|
|1|1|John|Doe|1000|1900-01-01|2023-01-01|false
|1|1|John|Doe|1200|2023-01-01|2023-01-01|false
|1|1|John|Doe|2000|2024-01-01|9999-01-01|false

### Delete record on 2024-06-01

|sid|id|first_name|last_name|salary|kvsz_start|kvsz_end|kvsz_deleted|
|---|--|----------|---------|------|----------|--------|------------|
|1|1|John|Doe|1000|1900-01-01|2023-01-01|false
|1|1|John|Doe|1200|2023-01-01|2023-01-01|false
|1|1|John|Doe|2000|2024-01-01|2024-06-01|true

## Source schema modifications

### Adding columns

If a column is added in a source database, it is ignored until it is added in the destination database. There is no automatic synchronization of columns. In most data consolidation scenarios, a subset of the source columns is required.

### Deleting columns

Columns should not be deleted from source tables. If they are deleted for any reason, they will be ognored in the destination table and the default value will be used. If the destination column does not allow NULLs and no default value is defined, the insert/update will fail.

### Changing column types

The destination column type should also be changed.

## Postgres configuration

- Configure replication slots in `postgresql.conf`

  ```
  max_replication_slots = 10
  max_wal_senders = 10`  -- there should be one slot for each replicated database plus one slot for each secondary server
  wal_level = logical
  ```

- Configure replication host in `pg_hba.conf` depending on where `kuvasz-streamer` is running.

  ```
  host    replication    all            0.0.0.0/0               scram-sha-256
  ```

- Create a replicaiton user exclusively for `kuvasz-streamer`

## Destination Schema

The following constraints apply to the destination schema

- Target tables can have a subset of the source tables
- Columns must have the same names and the same data types
- The target table has to have a `sid` column
- The target table primary key should be `tenand_id + source PK`

## Maintenance

The best way to monitor the replication state is to use `kuvasz-agent` and the associated Postgres Grafana dashboard.

- To add a new table to a replication set
    ```sql
    ALTER PUBLICATION kvsz_DBNAME  ADD TABLE TABLENAME;
    ```

- To check the replication slots
    ```sql
    SELECT *
    FROM pg_replication_slots;
    ```

- To check the replication status
    ```sql
    SELECT client_addr, state, sent_lsn write_lsn, flush_lsn, replay_lsn 
    FROM pg_stat_replication;
    ```

## Implementation details

The postgres replication protocol provides the necessary information to propagate the update. This table summarizes the various cases used to build a WHERE clause.

|Operation|Indicator|Values|OldValues|Where|Notes|
|---------|---------|------|---------|-----|-----|
|UPDATE|K|Modified values|Old primary Key|PK=OldValues.PK|Primary key was modified|
|UPDATE|O|Modified values|Full row|columns=OldValues|Replica identity full|
|UPDATE|00|Modified values including PK||PK=Values.PK|Primary Key was not modified|
|DELETE|K|Primary key||PK=Values.PK|Primary key was deleted|
|DELETE|O|Full row||columns=Values|Replica Identity Full
|DELETE|00|Should never happen|

## References

- Article
  - https://www.postgresql.fastware.com/blog/inside-logical-replication-in-postgresql
- Documentation
  - https://www.postgresql.org/docs/current/logical-replication.html
  - https://www.postgresql.org/docs/current/logicaldecoding.html
- Protocol: 
  - https://www.postgresql.org/docs/current/protocol-replication.html
  - https://www.postgresql.org/docs/current/protocol-logical-replication.html
- Go pgx and tools
  - https://github.com/jackc/pgx
  - https://github.com/jackc/pglogrepl
