# Getting Started

This document provides details to get you started with the development of the
Inventory extension for [Open Delivery Gear](https://github.com/open-component-model/ocm-gear).

The `inventory-extension-odg` extension is meant to be plugged into an existing
[gardener/inventory](https://github.com/gardener/inventory) cluster.

![Open Delivery Gear Extension](../images/inventory-extension-odg.png)

# Requirements

- Go 1.24.x or later
- [Redis](https://redis.io/)
- [PostgreSQL](https://www.postgresql.org/)
- [GNU Make](https://www.gnu.org/software/make/)

[Valkey](https://github.com/valkey-io/valkey) or [Redict](https://redict.io),
can be used instead of Redis.

Additional requirements for local development.

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

It is also recommended that you check the upstream
[gardener/inventory](https://github.com/gardener/inventory) documentation in
order to get familiar with the design and architecture of Inventory itself.

# Development Environment

A development environment can be started either locally, or in [Docker
Compose](https://docs.docker.com/compose/) using the provided
[docker-compose.yml](../docker-compose.yaml) manifest.

## Configuration

The included [examples/config.yaml](../examples/config.yaml) can be used as a
starting point to configure the extension. This configuration file can also be
specified via the `INVENTORY_EXTENSION_CONFIG` environment variable.

The `inventory-extension-odg` CLI app accepts multiple configuration files via
the `--config|--file <path>` option. This allows you to split the configuration
amongst multiple files, for better organization, if needed.

When specifying multiple configuration files via the
`INVENTORY_EXTENSION_CONFIG` env var, you need to separate the files using a
comma, e.g.

``` shell
env INVENTORY_EXTENSION_CONFIG=db.yaml,redis.yaml,odg.yaml /path/to/inventory-extension-odg worker start
```

## Database

The `inventory-extension-odg` extension requires a PostgreSQL database, which
must be configured via the [config file](../examples/config.yaml) provided to
the worker process.

This database is also used by Inventory itself, which populates data from the
supported datasources such as AWS, GCP, Azure, OpenStack, etc.

Unlike, upstream Inventory, which requires read-write access to the database,
the Open Delivery Gear extension needs read-only access only, so keep that in
mind when plugging this extension into an existing Inventory cluster.

In order to create a read-only user for the extension, when enabling it in an
existing Inventory cluster, you can use the following SQL statements against the
database used by Inventory itself.

``` sql
--
-- SQL script to create a read-only user for the inventory db
--
-- Make sure to specify the password in the first statement.
--
CREATE ROLE inventory_ro WITH LOGIN PASSWORD 'PASSWORD-GOES-HERE';
GRANT CONNECT ON DATABASE inventory TO inventory_ro;
GRANT USAGE ON SCHEMA public TO inventory_ro;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO inventory_ro;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO inventory_ro;
```

The SQL statements above will create a new PostgreSQL user `inventory_ro`, which
you can then use when configuring the extension.

If you need to test and develop the extension in isolation (meaning that no
upstream Inventory workers are present, but only the extension worker exists),
then you could simply take a backup of a database populated by Inventory,
restore it and configure the extension to use that.

## Standalone Mode

In order to start the extension on your local system you can run the following
command.

Note, that this command simply starts the extension worker process, which means
that you should already be running Redis and Postgres, and have the extension
configured appropriately against the respective endpoints.

``` shell
make run
```

Alternatively, you can build the worker extension binary and start it up
manually, e.g.

``` shell
make build
./bin/inventory-extension-odg worker start
```

## Docker Compose

You can also run a local development environment using Docker Compose, which
will start up Valkey, PostgreSQL and the extension worker for you. In order to do
that, simply execute the following command.

``` shell
make docker-compose-up
```

The services which will be started are summarized in the table below.

| Service    | Description                    |
|:-----------|:-------------------------------|
| `postgres` | PostgreSQL database            |
| `worker`   | Inventory Extension Worker     |
| `valkey`   | Valkey used as a message queue |

Once the services are up and running, you can access the following endpoints
from your local system.

| Endpoint               | Description                           |
|:-----------------------|:--------------------------------------|
| localhost:5432         | PostgreSQL server                     |
| localhost:6379         | Valkey server                         |
| localhost:6080/metrics | Metrics endpoint for Extension Worker |

If you want to run any additional upstream Inventory components such as the
`scheduler` or `dashboard`, please refer to the upstream
[Inventory Development Guide](https://github.com/gardener/inventory/blob/main/docs/development.md#docker-compose),
and the upstream [gardener/inventory docker-compose.yaml](https://github.com/gardener/inventory/blob/main/docker-compose.yaml)
manifest, which provides these services, as well as Grafana and Prometheus.

Additionally, you can extend the existing
[gardener/inventory docker-compose.yaml](https://github.com/gardener/inventory/blob/main/docker-compose.yaml),
manifest with the [gardener/inventory-extension-odg docker-compose.yaml](../docker-compose.yaml)
in order to run a complete Inventory cluster with the Open Delivery Gear extension in it.

# Metrics

The extension worker exposes the following metrics via it's metrics endpoint:

| Metric                                      | Type    | Description                                             |
|:--------------------------------------------|:--------|:--------------------------------------------------------|
| `inventory_odg_discovered_orphan_resources` | `gauge` | Number of discovered orphan resources from Inventory    |
| `inventory_odg_reported_orphan_resources`   | `gauge` | Number of successfully reported orphan resources to ODG |

`inventory-extension-odg` also exposes additional metrics provided by the
upstream [gardener/inventory](https://github.com/gardener/inventory), which
track successful/failed tasks, task execution duration, etc.

For more details about these metrics, please refer to the `gardener/inventory`
documentation.

# Extension Worker Tasks

The `inventory-extension-odg` extension provides the following tasks.

- `odg:task:report-orphan-vms-aws` - reports orphan AWS EC2 instances as findings
- `odg:task:report-orphan-vms-gcp` - reports orphan GCP Compute Engine instances as findings
- `odg:task:report-orphan-vms-az` - reports orphan Azure Virtual Machines as findings
- `odg:task:report-orphan-ip-addresses-gcp` - reports orphan GCP Public IP Addresses as findings

Each of these tasks expects a payload, which represents the query to be used
when fetching orphan resources from the database.

You can find example payloads in the [examples/payloads](../examples/payloads)
directory.

# Scheduler Jobs

Periodic jobs may be configured in the Inventory Scheduler, so that reporting on
orphan resources happens on periodic basis.

For more information about the Inventory Scheduler, please refer to the links
below.

- [Inventory Design](https://github.com/gardener/inventory/blob/main/docs/design.md)
- [Scheduler Component](https://github.com/gardener/inventory/blob/main/docs/development.md#scheduler)
- [Scheduler Config](https://github.com/gardener/inventory/blob/5ca666c0cfbbe5c0cec650f00d31811e772816fc/examples/config.yaml#L333)

You can find example scheduler jobs in the
[examples/scheduler](../examples/scheduler) directory.

# Tests

In order to run the test suite run the following command.

``` shell
make test
```
