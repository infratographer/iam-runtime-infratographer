{{- define "iam-runtime-infratographer.configmap" }}
{{- $values := (index .Subcharts "iam-runtime-infratographer").Values -}}
{{- $defaultConfig := dict "server" (dict "socketPath" "/var/iam-runtime/runtime.sock") }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "config" "context" $) | quote }}
  labels: {{- include "common.labels.standard" $ | nindent 4 }}
data:
  config.yaml: |
    {{- tpl (merge $defaultConfig $values.config | toYaml) $ | nindent 4 }}
{{- end }}
