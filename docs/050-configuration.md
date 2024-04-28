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
|`server`|`max_goroutines`|Integer|100|Number of concurrent API calls to process|
|`server`|`read_timeout`|Integer|30|Maximum time (in seconds) allowed to send the whole request|
|`server`|`read_header_timeout`|30|Integer|Maximum time (in seconds) allowed to send the header|
|`server`|`write_timeout`|Integer|30|Maximum time (in seconds) allowed to write the whole response|
|`server`|`idle_timeout`|Integer|30|Maximum time (in seconds) allowed between two requests on the same connection|
|`server`|`max_header_bytes`|Integer|1000|Maximum size (in bytes) of the headers|
|`maintenance`|`pprof`|String||Pprof bind adddress, typically `127.0.0.1:6060` when enabled|
|`maintenance`|`start_delay`|Integer|0|Testing only: delay between full sync and replication start|
|`database`|`url`|String||Destination database URL|
|`database`|`schema`|String|`public`|Destination database schema to use|
|`cors`|`allowed_origins`|Array of strings|Origin sites to allow, Use * for testing|
|`cors`|`allow_methods`|String|`GET,POST,PATCH,PUT,DELETE`|Comma separated list of allowed methods, should not be changed|
|`cors`|`allow_headers`|String|`Authorization,User-Agent,If-Modified-Since,Cache-Control,Content-Type,X-Total-Count`|Comma separated list of allowed headers, should not be changed|
|`cors`|`allow_credentials`|Boolean|true|Switch to allow Authorization header|
|`cors`|`max_age`|Integer|86400|Maximum time to use the CORS response in seconds|
|`auth`|`admin_password`|String|`hash(admin)`|Web administrator password. Compatible with `mkpasswd` output. |
|`auth`|`jwt_key`|String|`Y3OYHx7Y1KsRJPzJKqHGWfEaHsPbmwwSpPrXcND95Pw=`|JWT signing key. Generate a cryptographycally secure key with `openssl rand -base64 32`|
|`auth`|`ttl`|Integer|300|Token validity period in seconds|
|`app`|`map_file`|String|`map.yaml`|Table mapping file|
|`app`|`map_database`|String||Table mapping file|
|`app`|`num_workers`|Integer|2|Number of workers writing to the destination database|
|`app`|`commit_delay`|Float|1.0|Delay in seconds between commits on the destination database|
|`app`|`default_schema`|String|`public`|Default schema in source database|
|`app`|`sync_rate`|Float|1_000_000_000|Number of rows/second to read globally when doing a full sync in order not to overload the source database|
|`app`|`sync_burst`|Integer|1000|Number of rows to burst in case of delays in writing rows in the destination|


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