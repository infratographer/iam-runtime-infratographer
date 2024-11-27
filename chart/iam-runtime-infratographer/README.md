# iam-runtime-infratographer

Functions which assist in deploying iam-runtime-infratographer with your app.

## Example deployment

Helm chart repository: https://infratographer.github.io/charts

```yaml
# file: templates/deployment.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-app
spec:
  replicas: 1
  revisionHistoryLimit: 1
  selector:
    matchLabels:
      app: example-app
  template:
    metadata:
      annotations: {{- include "iam-runtime-infratographer.annotations" $ | nindent 8 }}
      labels:
        app: example-app
    spec:
      containers:
        - name: example-app
          image: {{ .Values.deployment.image }}
          volumeMounts: {{- include "iam-runtime-infratographer.volumeMounts" $ | nindent 12 }}
        - {{- include "iam-runtime-infratographer.container" $ | nindent 10 }}
      volumes: {{- include "iam-runtime-infratographer.volumes" $ | nindent 8 }}
{{- include "iam-runtime-infratographer.configmap" $ }}

# file: values.yaml
---
iam-runtime-infratographer:
  config:
    permissions:
      host: permissions-api.internal.example.net
    jwt:
      jwksURI: https://iam.example.com/jwks.json
      issuer: https://iam.example.com/
```

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | common | 2.22.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| config.accessTokenProvider.enabled | bool | `false` | enabled configures the access token source for GetAccessToken requests. |
| config.accessTokenProvider.exchange.grantType | string | urn:ietf:params:oauth:grant-type:token-exchange | grantType configures the grant type |
| config.accessTokenProvider.exchange.issuer | string | `""` | issuer specifies the URL for the issuer for the exchanged token. The Issuer must support OpenID discovery to discover the token endpoint. |
| config.accessTokenProvider.exchange.tokenType | string | urn:ietf:params:oauth:token-type:jwt | tokenType configures the token type |
| config.accessTokenProvider.expiryDelta | duration | 10s | expiryDelta sets early expiry validation for the token. |
| config.accessTokenProvider.source.clientCredentials.clientID | string | `""` | clientID is the client credentials id which is used to retrieve a token from the issuer. This attribute also supports a file path by prefixing the value with `file://`. example: `file:///var/secrets/client-id` |
| config.accessTokenProvider.source.clientCredentials.clientSecret | string | `""` | clientSecret is the client credentials secret which is used to retrieve a token from the issuer. This attribute also supports a file path by prefixing the value with `file://`. example: `file:///var/secrets/client-secret` |
| config.accessTokenProvider.source.clientCredentials.issuer | string | `""` | issuer specifies the URL for the issuer for the token request. The Issuer must support OpenID discovery to discover the token endpoint. |
| config.accessTokenProvider.source.file.tokenPath | string | `""` | tokenPath is the path to the source jwt token. |
| config.events.enabled | bool | `false` | enabled enables NATS event-based functions. |
| config.events.nats.credsFile | string | `""` | credsFile path to NATS credentials file |
| config.events.nats.publishPrefix | string | `""` | publishPrefix NATS publish prefix to use. |
| config.events.nats.publishTopic | string | `""` | publishTopic NATS publihs topic to use. |
| config.events.nats.token | string | `""` | token NATS user token to use. |
| config.events.nats.url | string | `""` | url NATS server url to use. |
| config.jwt.issuer | string | `""` | issuer Issuer to use for JWT validation. |
| config.jwt.jwksURI | string | `""` | jwksURI JWKS URI to use for JWT validation. |
| config.permissions.discovery.check.concurrency | int | `5` | concurrency is the number of hosts to concurrently check. |
| config.permissions.discovery.check.count | int | `5` | count is the number of checks to run on each host to check for connection latency. |
| config.permissions.discovery.check.delay | string | `"200ms"` | delay is the delay between requests for a host. |
| config.permissions.discovery.check.interval | string | `"1m"` | interval is how frequent to check for healthiness on hosts. |
| config.permissions.discovery.check.path | string | `"/readyz"` | path is the uri path to fetch to check if host is healthy. |
| config.permissions.discovery.check.scheme | string | `""` | scheme sets the uri scheme. Default is http unless discovered port is 443 in which https will be used. |
| config.permissions.discovery.check.timeout | string | `"2s"` | timeout sets the maximum amount of time a request can wait before canceling the request. |
| config.permissions.discovery.disable | bool | `false` | disable SRV discovery. |
| config.permissions.discovery.fallback | string | `""` | fallback sets the fallback address if no hosts are found or all hosts are unhealthy. The default fallback host is the permissions.host value. |
| config.permissions.discovery.interval | string | `"15m"` | interval to check for new SRV records. |
| config.permissions.discovery.optional | bool | `true` | optional allows SRV records to be optional. If no SRV records are found or all endpoints are unhealthy, the fallback host is used. |
| config.permissions.discovery.prefer | string | `""` | prefer sets the preferred SRV record. (skips priority, weight and duration ordering) |
| config.permissions.discovery.quick | bool | `false` | quick doesn't wait for discovery and health checks to complete before selecting a host. |
| config.permissions.host | string | `""` | host permissions-api host to use. |
| config.tracing.enabled | bool | `false` | enabled initializes otel tracing. |
| config.tracing.insecure | bool | `false` | insecure if TLS should be disabled. |
| config.tracing.url | string | `""` | url gRPC URL for OpenTelemetry collector. |
| extraEnv | object | `{}` | extraEnv defines additional environment variables to include with the container ref: https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/ |
| image.pullPolicy | string | `"IfNotPresent"` | pullPolicy is the image pull policy for the service image |
| image.repository | string | `"ghcr.io/infratographer/iam-runtime-infratographer"` | repository is the image repository to pull the image from |
| image.tag | string | `""` | tag is the image tag to use. Defaults to the chart's app version |
| resources | object | `{}` | resource limits & requests ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| restartPolicy | string | `""` | restartPolicy set to Always if using with initContainers on kube 1.29 and up with the SideContainer feature flag enabled. ref: https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/#sidecar-containers-and-pod-lifecycle |
| securityContext | object | `{"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true,"runAsNonRoot":true,"runAsUser":65532}` | securityContext configures the container's security context. ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| volumeMounts | object | `{}` | volumeMounts define additional volume mounts to include with the container ref: https://kubernetes.io/docs/concepts/storage/volumes/ |

