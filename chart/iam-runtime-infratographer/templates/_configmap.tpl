{{- define "iam-runtime-infratographer.configmap" }}
{{- $values := (index .Subcharts "iam-runtime-infratographer").Values -}}
{{- $defaultConfig := dict "server" (dict "socketPath" "/var/iam-runtime/runtime.sock") }}
{{- $config := include "iam-runtime-infratographer.omit" (dict
        "source" (merge $defaultConfig $values.config)
        "omit" (list
          "events.nats.token"
          "accessTokenProvider.source.clientCredentials.clientSecret"
        )
    )
}}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "config" "context" $) | quote }}
  labels: {{- include "common.labels.standard" $ | nindent 4 }}
data:
  config.yaml: |
    {{- tpl $config $ | nindent 4 }}
{{- end }}
