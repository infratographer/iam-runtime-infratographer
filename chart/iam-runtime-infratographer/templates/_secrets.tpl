{{- define "iam-runtime-infratographer.secrets" }}
{{- $values := (index .Subcharts "iam-runtime-infratographer").Values -}}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "secrets" "context" $) | quote }}
  labels: {{- include "common.labels.standard" $ | nindent 4 }}
data:
  natsToken: {{ $values.secrets.nats.token | quote }}
  clientSecret: {{ $values.secrets.accessToken.source.clientSecret | quote }}
{{- end }}
