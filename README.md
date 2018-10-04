# Metric Registrar CLI
Plugin for the CF CLI that allows users to register metric sources for collection.

## Installing Plugin
`cf install-plugin -r CF-Community "metric-registrar"`

## Usage
There are two types of registries

### Structured Log Format
Registering a structured log format will allow for structured logs of that format to be parsed into metrics and events and emitted to Loggregator.

```
cf register-log-format --help
NAME:
   register-log-format - This will register bound applications so that structured logs of the given format can be parsed

USAGE:
   cf register-log-format APPNAME FORMAT
```
FORMAT should be either `json` or `DogStatsD`

### Metrics Endpoint
Registering a metrics endpoint will allow for Prometheus Exposition metrics from the given path to be parsed and emitted to Loggregator.

```
cf register-metrics-endpoint --help
NAME:
   register-metrics-endpoint - This will register your metrics endpoint which will then be scraped at the interval defined at deploy

USAGE:
   cf register-metrics-endpoint APPNAME PATH
```
