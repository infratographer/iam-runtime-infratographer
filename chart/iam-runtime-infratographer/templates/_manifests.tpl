{{- define "iam-runtime-infratographer.manifests" }}
{{ include "iam-runtime-infratographer._configmap" $ }}
{{ include "iam-runtime-infratographer._secrets" $ }}
{{- end }}
