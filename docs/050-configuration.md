---
layout: page
title: Configuration
permalink: /configuration/
nav_order: 50
---
# Configuration

## Service configuration

`kuvasz-streamer` supports three configuration sources with the following order of priority.

1. Configuration file with TOML syntax. If the file name is not specified with the 
   `--conf` command line switch, it searches for the following three files in order:
  - `kuvasz-streamer.toml`
  - `./conf/kuvasz-streamer.toml`
  - `/etc/kuvasz/kuvasz-streamer.toml`

2. Environment variables starting with the format `KUVASZ_section_parameter=value`

3. Command line arguments in the form `--section.parameters=value`

|Section|Parameter|Type|Default|Description|
|-------|---------|----|-------|-----------|
|`server`|`name`|String|`kuvasz-streamer`|Server name to use in log shipping|
|`server`|`address`|String|:8000|Server bind address|
|`server`|`pprof`|String||Pprof bind adddress, typically `127.0.0.1:6060` when enabled|
|`server`|`start_delay`|Integer|0|Testing only: delay between full sync and replication start|
|`database`|`url`|String||Destination database URL|
|`app`|`map_file`|String|`map.yaml`|Table mapping file|

## Mapping file

The mapping file is a YAML formatted file that maps the source databases and tables to the destinations. For each source database schema, create a top-level key with an identifier. Then list all the URLs to access engines with this schema and for each engine, specify a source ID `sid` to differentiate in the destination database.

```yaml
# Top level key for all databases with the same schema
- database: db1
  urls:
  - url: postgres://kuvasz:kuvasz@127.0.0.1:6012/db1?replication=database&application_name=repl_db1
    sid: i1  # identifier in the destination database
  - url: postgres://kuvasz:kuvasz@127.0.0.1:6013/db1?replication=database&application_name=repl_db1
    sid: i2
  # List all tables to be replicated
  tables:
    t1:
    t2:
      target: rt2 # Table name in destination database
    t3:
      type: append # Specify table type append
    t4:
      type: history # specify table type history
- database: db2
  urls:
  - url: postgres://kuvasz:kuvasz@127.0.0.1:6012/db2?replication=database&application_name=repl_db2
    sid: 12
  - url: postgres://kuvasz:kuvasz@127.0.0.1:6013/db2?replication=database&application_name=repl_db2
    sid: 13
  tables:
    s1:
```