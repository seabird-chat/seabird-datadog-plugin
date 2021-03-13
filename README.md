# seabird-datadog-plugin
Seabird plugin to send chat metrics to Datadog

Inspired by [pisg](http://pisg.sourceforge.net/), this plugin generates 
metrics from messages sent to IRC rooms that seabird is in. 


## Requirements
* Running `datadog-agent` progress with `dogstatsd` enabled
* Environment variables `SEABIRD_HOST`, `SEABIRD_TOKEN`, and `DOGSTATSD_ENDPOINT` configured