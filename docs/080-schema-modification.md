---
layout: page
title: Schema Modification
permalink: /schema-modification/
nav_order: 80
---
# Schema Modification

## Adding columns

If a column is added in a source database, it is ignored until it is added in the destination database. There is no automatic synchronization of columns. In most data consolidation scenarios, a subset of the source columns is required.

## Deleting columns

Columns should not be deleted from source tables. If they are deleted for any reason, they will be ognored in the destination table and the default value will be used. If the destination column does not allow NULLs and no default value is defined, the insert/update will fail.

## Changing column types

The destination column type should also be changed.

