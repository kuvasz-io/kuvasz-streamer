---
layout: page
title: Postgres configuration
permalink: /postgres-configuration/
nav_order: 60
---
# Postgres Configuration

## Postgres server configuration

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
- The target table primary key should be `sid + source PK`

