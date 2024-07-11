# Changelog

## v2.1.1

- add help message to the linux installation configuration

## v2.1.0

- Update dependencies
- add linux package

## v2.0.10

- Update dependencies (go 1.22)

## v2.0.9

- Update dependencies

## v2.0.8

- Update dependencies

## v2.0.7

- Update dependencies

## v2.0.6

- Update dependencies

## v2.0.5

- Added `pprof` endpoints for debugging purposes
- Update dependencies

## v2.0.3

- Possibility to exclude attributes from discovery

## v2.0.2

- update dependencies

## v2.0.2

- migration to new unified steadybit actionIds and targetTypes

## v2.0.1

- update dependencies

## v2.0.0

- Refactoring using `action_kit_sdk`
- Read only file system

## v1.4.0

 - Print build information on extension startup.

## v1.3.0

 - Support creation of a TLS server through the environment variables `STEADYBIT_EXTENSION_TLS_SERVER_CERT` and `STEADYBIT_EXTENSION_TLS_SERVER_KEY`. Both environment variables must refer to files containing the certificate and key in PEM format.
 - Support mutual TLS through the environment variable `STEADYBIT_EXTENSION_TLS_CLIENT_CAS`. The environment must refer to a comma-separated list of files containing allowed clients' CA certificates in PEM format.

## v1.2.0

- Support for the `STEADYBIT_LOG_FORMAT` env variable. When set to `json`, extensions will log JSON lines to stderr.

## v1.1.1

 - Set category to `monitoring` to align with Datadog extension.

## v1.1.0

 - The log level can now be configured through the `STEADYBIT_LOG_LEVEL` environment variable.

## v1.0.1

 - update ActionKit API to fix `executionId` deserialization

## v1.0.0

 - Initial release
