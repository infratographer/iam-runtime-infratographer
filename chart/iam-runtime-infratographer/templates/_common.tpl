{{- define "iam-runtime-infratographer.fullname" }}
  {{- $values := (index .Subcharts "iam-runtime-infratographer").Values -}}
  {{- include "common.names.dependency.fullname" (dict "chartName" "iam-runtime-infratographer" "chartValues" $values "context" $) -}}
{{- end }}

{{- define "iam-runtime-infratographer.resource.fullname" }}
  {{- $prefix := include "iam-runtime-infratographer.fullname" .context }}
  {{- $totalLength := add (len $prefix) (len .suffix) 1 | int }}
  {{- $trimLength := sub $totalLength 63 | int }}
  {{- $prefix = trunc (sub (len $prefix) $trimLength | int) $prefix | trimSuffix "-" }}
  {{- printf "%s-%s" $prefix .suffix -}}
{{- end }}

{{- define "iam-runtime-infratographer.annotations" -}}
checksum/iam-runtime-infratographer-config: {{ toYaml (index .Subcharts "iam-runtime-infratographer").Values | sha256sum }}
{{- end }}
