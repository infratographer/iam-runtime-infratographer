{{- define "iam-runtime-infratographer.volumes" -}}
{{- $values := (index .Subcharts "iam-runtime-infratographer").Values -}}
- name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "config" "context" $) | quote }}
  configMap:
    name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "config" "context" $) | quote }}
- name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "socket" "context" $) | quote }}
  emptyDir: {}
{{- end }}

{{- define "iam-runtime-infratographer.volumeMounts" -}}
- name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "socket" "context" $) | quote }}
  mountPath: {{ .Values.socketVolumeMountPath }}
{{- end }}
