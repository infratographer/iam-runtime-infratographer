{{- define "iam-runtime-infratographer.fullname" }}
  {{- $values := (index .Subcharts "iam-runtime-infratographer").Values -}}
  {{- include "common.names.dependency.fullname" (dict "chartName" "iam-runtime-infratographer" "chartValues" $values "context" $) -}}
{{- end }}

{{- define "iam-runtime-infratographer.resource.fullname" }}
  {{- $prefix := include "iam-runtime-infratographer.fullname" .context }}
  {{- $reduce := sub (add (len $prefix) (len .suffix) 1) 63 }}
  {{- if gt $reduce 0 }}
    {{- $prefix = trunc (add 63 $reduce) $prefix | trimSuffix "-" }}
  {{- end }}
  {{- printf "%s-%s" $prefix .suffix -}}
{{- end }}

{{- define "iam-runtime-infratographer.annotations" -}}
checksum/iam-runtime-infratographer-config: {{ toYaml (index .Subcharts "iam-runtime-infratographer").Values | sha256sum }}
{{- end }}
