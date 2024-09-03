{{- define "iam-runtime-infratographer.secrets" }}
{{- $values := (index .Subcharts "iam-runtime-infratographer").Values -}}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "secrets" "context" $) | quote }}
  labels: {{- include "common.labels.standard" $ | nindent 4 }}
data:
  {{- with $values.config.events.nats.token }}
  IAMRUNTIME_EVENTS_NATS_TOKEN: {{ quote . }}
  {{- end }}
  {{- with $values.config.accessTokenProvider.source.clientCredentials.clientSecret }}
  IAMRUNTIME_ACCESSTOKENPROVIDER_SOURCE_CLIENTCREDENTIALS_CLIENTSECRET: {{ quote . }}
  {{- end }}
{{- end }}
