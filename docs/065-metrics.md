---
layout: page
title: Metrics
permalink: /metrics/
nav_order: 65
---
# Metrics

`kuvasz-steramer` maintains Prometheus metrics available on `/metrics` endpoint.

|Metric|Type|Labels|Description|
|------|----|------|-----------|
|`streamer_operations_total`|Counter|`database`, `sid`, `table`, `operation`, `result`|Total number of INSERT/UPDATE/DELETE operations|
|`streamer_operations_seconds`|Histogram|`database`, `sid`, `table`, `operation`, `result`|Duration of INSERT/UPDATE/DELETE operations|
|`streamer_sync_total_rows`|Counter|`database`, `sid`, `table`|Total number of rows synced|
|`streamer_sync_total_bytes`|Counter|`database`, `sid`, `table`|Total number of bytes synced|
|`streamer_jobs_total`|Counter|`channel`|Total number of jobs received per channel|
|`url_heartbeat`|Gauge|`database`,`sid`|Timestamp of last known activity|
