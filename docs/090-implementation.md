---
layout: page
title: Implementation details
permalink: /implementation/
nav_order: 90
---
# Implementation details

## Postgres replication protol use-cases

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
  - [Fastware](https://www.postgresql.fastware.com/blog/inside-logical-replication-in-postgresql)
- Postgres Documentation
  - [Logical replication](https://www.postgresql.org/docs/current/logical-replication.html)
  - [Logical decoding](https://www.postgresql.org/docs/current/logicaldecoding.html)
- Protocol: 
  - [Streaming Replication](https://www.postgresql.org/docs/current/protocol-replication.html)
  - [Logical Replication](https://www.postgresql.org/docs/current/protocol-logical-replication.html)
- Go pgx and tools
  - [jackc/pgx](https://github.com/jackc/pgx)
  - [jack/pglogrepl](https://github.com/jackc/pglogrepl)
