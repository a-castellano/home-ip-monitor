# Home IP Monitor

[![pipeline status](https://git.windmaker.net/a-castellano/home-ip-monitor/badges/master/pipeline.svg)](https://git.windmaker.net/a-castellano/home-ip-monitor/pipelines)[![coverage report](https://git.windmaker.net/a-castellano/home-ip-monitor/badges/master/coverage.svg)](https://a-castellano.gitpages.windmaker.net/home-ip-monitor/coverage.html)[![Quality Gate Status](https://sonarqube.windmaker.net/api/project_badges/measure?project=home-ip-monitor&metric=alert_status)](https://sonarqube.windmaker.net/dashboard?id=home-ip-monitor)

This program checks if my Home IP changes, notify if IP and ISP changes.

# What this progam does?

This program fetch my public IP, checks if it has changed since last check.
* If IP has changed and it belongs to required ISP, IP is updated and notified.
* If new IP belogns to different ISP it only notifies
* Otherwise this program does nothing.


# Required variables

## ISP variables

**ISP_NAME**: Name of connection ISP current copany, value must be the same as the one showed [https://ipinfo.io/](https://ipinfo.io/) **ASN** field, for example "DIGI".

## Queue names

**UPDATE_QUEUE_NAME**: Queue name where new IP's will be sended.

**NOTIFY_QUEUE_NAME**: Queue name where notifications will be sended.

## Redis Config

Redis required config can be found in its [go types](https://git.windmaker.net/a-castellano/go-types/-/tree/master/redis?ref_type=heads) Readme.

## RabbitMQ Config

RabbitMQ required config can be found in its [go types](https://git.windmaker.net/a-castellano/go-types/-/tree/master/rabbitmq?ref_type=heads) Readme.
