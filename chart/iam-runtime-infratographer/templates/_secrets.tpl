{{- define "iam-runtime-infratographer.secrets" }}
{{- $values := (index .Subcharts "iam-runtime-infratographer").Values -}}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "secrets" "context" $) | quote }}
  labels: {{- include "common.labels.standard" $ | nindent 4 }}
data:
  IAMRUNTIME_EVENTS_NATS_TOKEN: {{ $values.secrets.nats.token | quote }}
  IAMRUNTIME_ACCESSTOKENPROVIDER_SOURCE_CLIENTCREDENTIALS_CLIENTSECRET: {{ $values.secrets.accessToken.source.clientSecret | quote }}
{{- end }}
