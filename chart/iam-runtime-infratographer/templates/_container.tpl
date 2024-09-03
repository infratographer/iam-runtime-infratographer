{{- define "iam-runtime-infratographer.container" -}}
{{- $values := (index .Subcharts "iam-runtime-infratographer").Values -}}
name: {{ include "iam-runtime-infratographer.fullname" $ | quote }}
image: {{ include "iam-runtime-infratographer.container.image" $ | quote }}
imagePullPolicy: {{ quote $values.image.pullPolicy }}
{{- with $values.restartPolicy }}
restartPolicy: {{ quote . }}
{{- end }}
{{- with $values.securityContext }}
securityContext: {{- toYaml . | nindent 2 }}
{{- end }}
{{- with $values.resources }}
resources: {{- toYaml . | nindent 2 }}
{{- end }}
envFrom:
  - secretRef:
      name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "secrets" "context" $) | quote }}
{{- with $values.extraEnv }}
env:
 {{- toYaml . | nindent 2 }}
{{- end }}
volumeMounts:
  - name: {{ include "iam-runtime-infratographer.resource.fullname" (dict "suffix" "config" "context" $) | quote }}
    mountPath: /etc/iam-runtime-infratographer/
  {{- include "iam-runtime-infratographer.volumeMounts" $ | nindent 2 }}
  {{- with $values.volumeMounts }}
    {{- toYaml . | nindent 2 }}
  {{- end }}
{{- end }}

{{- define "iam-runtime-infratographer.container.image" }}
  {{- $values := (index .Subcharts "iam-runtime-infratographer").Values -}}
  {{- $tag := default (index .Subcharts "iam-runtime-infratographer" "Chart").AppVersion $values.image.tag }}
  {{- printf "%s:%s" $values.image.repository $tag }}
{{- end }}
