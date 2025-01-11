---
layout: page
title: Use cases
permalink: /use-cases/
nav_order: 10
---
# Use cases

## 1. Microservice database consolidation
![Consolidation](/assets/images/consolidation.png)

In a microservices architecture, each service has its own database. This poses a number of problems for reporting:
- Performing reporting queries on a production database will slow it down and make access times unpredictable.
- Indexes required for reporting add an overhead on the insert/update/delete operations.
- It is not possible to query across multiple databases.

`kuvasz-streamer` consolidates all the databases of all services into a single data warehouse. Only the required tables and columns are replicated.

The schema in the data warehouse does not have to follow the same one as the original services.

## 2. Multi-tenant database consolidation
![multinenant](/assets/images/multitenant.png)

In a sensitive multi-tenant environment, each tenant is assigned a separate database to ensure that no cross-pollination of data occurs. However, all the tenant databases have the exact same schema.

This poses a problem for cross-tenant reporting and customer support. `kuvasz-streamer` can be used to consolidate all the data in a single database. A newly added column `sid` identifies the source tenant database.

## 3. Database performance optimization
![optimize](/assets/images/optimize.png)
In a high-performance system, historical data is kept to a minimum in order to provide quick query time and low latency to end users. A cleaner process usually deletes all data older than a certain number of weeks. If tables are partitioned, old paritions are dropped.

However, historical data is important for, AI/ML, reporting and forensics. `kuvasz-streamer` can be configured to ignore the `DELETE` and `TRUNCATE` operations on some tables and only apply the `INSERT` and `UPDATE` operations. A separate cleaner running on the data warehouse database takes care of removing older historical data.

Example usage includes transaction tables and audit history tables.

## 4. Postgres major version upgrade
![upgrade](/assets/images/upgrade.png)
Upgrading major versions of Postgres is a time-consuming task that requires substantial downtime. `kuvasz-streamer` can be used to synchronize databases between different versions of Postgres and then performing a quick switchover.