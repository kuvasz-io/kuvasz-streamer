---
layout: page
title: Maintenance
permalink: /maintenance/
nav_order: 70
---
# Maintenance

The best way to monitor the replication state is to use `kuvasz-agent` and the associated Postgres Grafana dashboard.

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
