# inventory-extension-odg

`inventory-extension-odg` is an extension for
[gardener/inventory](https://github.com/gardener/inventory), which provides
integration with [Open Delivery Gear](https://github.com/open-component-model/ocm-gear).

Orphan resources discovered by Inventory will be submitted by the
`inventory-extension-odg` extension to the Delivery Service API as findings.

The following diagram provides a high-level overview of how the extension plugs
into the existing [gardener/inventory](https://github.com/gardener/inventory)
architecture.

![Open Delivery Gear Extension](./images/inventory-extension-odg.png)

# Documentation

Check the [Getting Started](./docs/getting-started.md) documentation for details
on how to get started with `inventory-extension-odg`, how to configure the
extension, setup a local development environment, etc.

# License

This project is Open Source and licensed under [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
