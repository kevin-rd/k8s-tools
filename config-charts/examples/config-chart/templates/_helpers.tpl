{{/*
Expand the chart name.
*/}}
{{- define "demo-app-config.name" -}}
{{- default .Release.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Return the ConfigMap name that the app chart references through fileConfig.existingName.
*/}}
{{- define "demo-app-config.configMapName" -}}
{{- if .Values.configMap.name -}}
{{- .Values.configMap.name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "demo-app-config.name" . -}}
{{- if hasSuffix "-config" $name -}}
{{- $name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-config" $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "demo-app-config.labels" -}}
app.kubernetes.io/name: {{ include "demo-app-config.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
