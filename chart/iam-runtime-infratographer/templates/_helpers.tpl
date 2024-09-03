{{- define "iam-runtime-infratographer.omit" }}
  {{- $subOmit := list }}
  {{- range .omit }}
    {{- if contains "." . }}
      {{- $subkey := splitList "." . | rest | join "." }}
      {{- $subOmit = append $subOmit $subkey }}
    {{- end}}
  {{- end }}

  {{- $result := dict }}
  {{- range $key, $val := .source }}
    {{- if has $key $.omit }}
      {{- /* key is ommited */}}
    {{- else if and $subOmit (kindIs "map" $val) }}
      {{- $ctx := dict
            "source" $val
            "omit" $subOmit
            "quiet" true
      }}
      {{- include "iam-runtime-infratographer.omit" $ctx }}
      {{- $_ := set $result $key $ctx.source }}
    {{- else }}
      {{- $_ := set $result $key $val }}
    {{- end }}
  {{- end }}

  {{- $_ := set . "source" $result }}

  {{- if not .quiet }}
    {{- toYaml $result }}
  {{- end }}
{{- end }}
