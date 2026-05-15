{{/*
Expand the chart name.
*/}}
{{- define "config-chart.name" -}}
{{- default .Release.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Return the ConfigMap name that the app chart references through fileConfig.existingName.
*/}}
{{- define "config-chart.configMapName" -}}
{{- $configMap := default dict .Values.configMap -}}
{{- if $configMap.name -}}
{{- $configMap.name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "config-chart.name" . -}}
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
{{- define "config-chart.labels" -}}
app.kubernetes.io/name: {{ include "config-chart.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
