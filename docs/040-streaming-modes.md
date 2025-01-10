---
layout: page
title: Streaming modes
permalink: /streaming-modes/
nav_order: 40
---

# Streaming modes

The streaming mode can be defined per table and affects how replication operations are applied on the destination. This is one of the main differences with normal Postgres logical replication where only exact copies are supported.

## type = `clone`
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

## type = `append`
In high-performance systems, it is important to keep a small number of historical events in the live service table and keep a much larger history in a data warfehouse. Examples: audit events, notifications, transactions. 

These tables behave the same as `clone` tables with the exception that `DELETE` and `TRUNCATE` are ignored.

## type = `history`
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
|1|1|John|Doe|1000|1900-01-01|2023-01-01|false
|1|1|John|Doe|1200|2023-01-01|9999-01-01|false

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
