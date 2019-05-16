# Metrics on each host

DNS discovery + URL with `lookup=docker`

# Metrics from each service instance

URL to proxy service
Result relabeled with `container=stack-service-N-XXXXXXXX`

# URL format
  `https://proxy.prometheus:9999/app.surveda:12313/metrics?basic_auth=metrics:`

# Lookup
  * dns (ip)
  * docker (container)
  * rancher (container)
  * kubernetes (pod)
