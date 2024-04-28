---
layout: page
title: Running modes
permalink: /running-modes/
nav_order: 45
---

# Running modes

`kuvasz-streamer` supports two running modes, each suitable for a different environment.

## Declarative mode
In this mode, the mapping configuration (databases, URLs, mappings) is statically configured in a YAML file. Any modification of the file requires a restart of the service.

This mode is suitable in Kubernetes clusters where the streamer reads its configuration from file generated in GitOps CI/CD pipelines. It does not require any mounted ephemeral or persistent storage. The web administration and APIs runs in read-only mode.

This mode is enabled when no database is specified in the configuration, ie when `app.map_database` is empty.

## Database mode
In database mode, the streamer requires a persistent read/write storage for an SQLite database containing its mapping configuration. This allows the administrator to add and remove databases and mappings at runtime and call an API to refresh the configuration.

This mode is suitable when running as a system service and experimentation with various mappings is desired. It is enabled by specifying the SQLite database path. All schema migrations are handled transparently by the service.
 

