{{- define "iam-runtime-infratographer.manifests" }}
{{ include "iam-runtime-infratographer.configmap" $ }}
{{ include "iam-runtime-infratographer.secrets" $ }}
{{- end }}
