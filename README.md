# Metric Registrar CLI
Plugin for the CF CLI that allows users to register metric sources for collection.

## Installing Plugin
`cf install-plugin -r CF-Community "metric-registrar"`

## Usage
There are two types of registries

### Metrics Endpoint
Registering a metrics endpoint will allow for Prometheus Exposition metrics from the given path to be parsed and emitted to Loggregator.

```
cf register-metrics-endpoint --help
NAME:
   register-metrics-endpoint - Register a metrics endpoint which will be scraped at the interval defined at deploy

USAGE:
   cf register-metrics-endpoint APPNAME PATH
```

### Structured Log Format
Registering a structured log format will allow for structured logs of that format to be parsed into metrics and events and emitted to Loggregator.

```
cf register-log-format --help
NAME:
   register-log-format - Register bound applications so that structured logs of the given format can be parsed

USAGE:
   cf register-log-format APPNAME <json|DogStatsD>
```

## Supported Log Structures

#### JSON
  - **Events**
    ```
    {
      "type": "event",
      "title": "title",
      "body": "body",
      "tags": {
        "tag1": "tag value"
      }
    }
    ```
    Translated to [Loggregator Events][loggregator-event].

 - **Gauges**
    ```
    {
      "type": "gauge",
      "name": "some-counter",
      "value": <float>,
      "tags": {
        "tag1": "tag value"
      }
    }
    ```
    Translated to [Loggregator Gauges][loggregator-gauge].

 - **Counters**
    ```
    {
      "type": "counter",
      "name": "some-counter",
      "delta": <uint>,
      "tags": {
        "tag1": "tag value"
      }
    }
    ```
    Translated to [Loggregator Counters][loggregator-counter].

#### [dogstatsd](dogstatsd-spec)

 - **Events**
   `_e{title.length,text.length}:title|text|d:timestamp|h:hostname|p:priority|t:alert_type|#tag1,tag2`
   
   Translated to [Loggregator Events][loggregator-event].
   
   *Supported Fields:*
   - `title`
   - `text`
   
   *Not Supported:*
   - `timestamp` - Log message timestamp is used
   - `hostname` - Source and Instance IDs from log message are used
   - `priority`
   - `alert_type`
   - arbitrary tags
   
 - **Gauges**
   `gauge.name:value|g|@sample_rate|#tag1:value,tag2`
  
   Translated to [Loggregator Gauges][loggregator-gauge].
    
    *Supported Fields:*
    - `gauge.name`
    - `value`
    
    *Not Supported:*
    - `sample_rate`
    - arbitrary tags

 - **Counters**
   `counter.name:value|c|@sample_rate|#tag1:value,tag2`
  
    Translated to [Loggregator Counters][loggregator-counter].
    
    *Supported Fields:*
    - `counter.name`
    - `value`
    
    *Not Supported:*
    - `sample_rate`
    - arbitrary tags


