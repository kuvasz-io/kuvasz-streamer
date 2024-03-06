---
layout: page
title: Getting started
permalink: /getting-started/
nav_order: 30
---
# Getting started

This guide runs a source and destination instance in Docker and `kuvasz-streamer` as a system service. It assumes running udner Ubunti 22.04 LTS

## Make sure Docker and Postgres are installed on the host

```bash
sudo apt install docker.io postgresql postgresql-contrib
```

## Start source database on port 6015 and create source schema

Run the following in a first window

```bash
sudo docker pull postgres:15
sudo docker run -i -t --name source -p 6015:5432 -e POSTGRES_PASSWORD=postgres postgres:15 -c wal_level=logical -c log_connections=on -c log_min_duration_statement=0
psql postgres://postgres:postgres@127.0.0.1:6015/postgres -c "create database source"
psql postgres://postgres:postgres@127.0.0.1:6015/source -c "create table employee(id serial, name text, dob date, salary numeric)"
psql postgres://postgres:postgres@127.0.0.1:6015/source -c "create publication kuvasz_source on all tables"
```

## Start destination database on port 6016 and create destination schema

Run this in a second window

```bash
sudo docker pull postgres:16
sudo docker run -i -t --name dest -p 6016:5432 -e POSTGRES_PASSWORD=postgres postgres:16 -c log_connections=on -c log_min_duration_statement=0
psql postgres://postgres:postgres@127.0.0.1:6016/postgres -c "create database dest"
psql postgres://postgres:postgres@127.0.0.1:6016/dest -c "create table emp(sid text, id int, name text, dob date)"
```

## Configure streamer

In a third window, create streamer config file with minimal configuration

```bash
cat <<EOF > kuvasz-streamer.toml
[database]
url = "postgres://postgres:postgres@127.0.0.1:6016/dest?application_name=kuvasz-streamer"
[app]
map_file = "./map.yaml"
EOF
```

Create map file

```bash
cat <<EOF > map.yaml
- database: source
  urls:
  - url: postgres://postgres:postgres@127.0.0.1:6015/source?replication=database&application_name=repl_source
    sid: source
  tables:
    employee:
      target: emp
      type: append
EOF
```

Start the streamer

```bash
kuvasz-streamer --conf ./kuvasz-streamer.toml
```

## Test

In a fourth window, insert a record in the source database

```bash
psql postgres://postgres:postgres@127.0.0.1:6015/source -c "insert into employee(name, dob, salary) values('tata', '1970-01-02', 2000)"
```

Now check it has been replicated to the destination database

```bash
psql postgres://postgres:postgres@127.0.0.1:6016/source -c "select * from emp"
```
