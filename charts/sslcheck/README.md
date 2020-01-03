# SSL Checker

Helm Chart to run the SSL Checker application as a Kubernetes CronJob.

## Installation

```
helm upgrade --install \
	sslchecker \
	./charts/sslcheck
```

## Configuration

| Parameter                     | Description                            | Default         |
|-------------------------------|----------------------------------------|-----------------|
| `image.repository`            | Container image to deploy              | `soon/sslcheck` |
| `image.pullPolicy`            | Image pull policy                      | `IfNotPresent`  |
| `schedule`                    | Cron schedule to trigger job           | `0 10 * * 3`    |
| `slackHookUrl`                | Slack webhook notifcation (optional)   | ``              |
| `sslExpiryThreshold.warning`  | Threshold in days for warning level    | `30`            |
| `sslExpiryThreshold.critical` | Threshold in days for critical level   | `14`            |
| `hosts`                       | List of hosts to check SSL certificate | `[]`            |
| `imagePullSecrets`            |                                        | `{}`            |
| `nameOverride`                |                                        | ``              |
| `fullnameOverride`            |                                        | ``              |
| `serviceAccount.create`       | Create service account                 | `true`          |
| `serviceAccount.name`         | Override service account name          | ``              |
| `podSecurityContext`          |                                        | `{}`            |
| `securityContext`             |                                        | `{}`            |
| `resources`                   | Resource allocation (YAML)             | `{}`            |
| `nodeSelector`                |                                        | `{}`            |
| `tolerations`                 |                                        | `[]`            |
| `affinity`                    |                                        | `{}`            |
