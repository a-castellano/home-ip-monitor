# Home IP Monitor

[![pipeline status](https://git.windmaker.net/a-castellano/home-ip-monitor/badges/master/pipeline.svg)](https://git.windmaker.net/a-castellano/home-ip-monitor/pipelines)[![coverage report](https://git.windmaker.net/a-castellano/home-ip-monitor/badges/master/coverage.svg)](https://a-castellano.gitpages.windmaker.net/home-ip-monitor/coverage.html)[![Quality Gate Status](https://sonarqube.windmaker.net/api/project_badges/measure?project=a-castellano_home-ip-monitor_a0d9946c-4181-4181-af10-e5dac69d0658&metric=alert_status&token=sqb_991ee37d1ea08ee63db5ea610f2a2d9e49fe1430)](https://sonarqube.windmaker.net/dashboard?id=a-castellano_home-ip-monitor_a0d9946c-4181-4181-af10-e5dac69d0658)

This program checks if my Home IP changes, notify if IP and ISP changes.

# What this program does?

This program fetch my public IP, checks if it has changed since last check.

- If IP has changed and it belongs to required ISP, IP is updated and notified.
- If new IP belongs to different ISP it only notifies
- Otherwise this program does nothing.

# Required variables

## ISP variables

**ISP_NAME**: Name of connection ISP current company, value must be the same as the one shown [https://ipinfo.io/](https://ipinfo.io/) **ASN** field, for example "DIGI".

## Queue names

**UPDATE_QUEUE_NAME**: Queue name where new IP's will be sent.

**NOTIFY_QUEUE_NAME**: Queue name where notifications will be sent.

## Redis Config

Redis required config can be found in its [go types](https://git.windmaker.net/a-castellano/go-types/-/tree/master/redis?ref_type=heads) Readme.

## RabbitMQ Config

RabbitMQ required config can be found in its [go types](https://git.windmaker.net/a-castellano/go-types/-/tree/master/rabbitmq?ref_type=heads) Readme.
