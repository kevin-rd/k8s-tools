{{/*
Expand the name of the chart.
*/}}
{{- define "generic.name" -}}
{{- default .Release.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "generic.fullname" -}}
{{- default .Release.Name .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "generic.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "generic.labels" -}}
{{ include "generic.selectorLabels" . }}
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "generic.selectorLabels" -}}
app.kubernetes.io/name: {{ include "generic.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use.
*/}}
{{- define "generic.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "generic.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the primary Service port used by single-backend resources.
*/}}
{{- define "generic.primaryServicePort" -}}
{{- $firstPort := first (required "ports must contain at least one port" .Values.ports) -}}
{{- required "ports[0].port is required" (get $firstPort "port") -}}
{{- end }}

{{/*
Return the claim name for a persistence item.
*/}}
{{- define "generic.persistenceClaimName" -}}
{{- $root := index . 0 -}}
{{- $claim := index . 1 -}}
{{- default (printf "%s-%s" (include "generic.fullname" $root) ($claim.name | default "data" | lower)) $claim.existingClaim | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Return the ConfigMap name used for envFrom.
*/}}
{{- define "generic.envConfigName" -}}
{{- if .Values.envConfig.existingName -}}
{{- .Values.envConfig.existingName -}}
{{- else -}}
{{- if empty .Values.envConfig.data -}}
{{- fail "envConfig.data is required when envConfig.enabled=true and envConfig.existingName is empty" -}}
{{- end -}}
{{- printf "%s-env-config" (include "generic.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end }}

{{/*
Return the Secret name used for envFrom.
*/}}
{{- define "generic.envSecretName" -}}
{{- if .Values.envSecret.existingName -}}
{{- .Values.envSecret.existingName -}}
{{- else -}}
{{- if empty .Values.envSecret.stringData -}}
{{- fail "envSecret.stringData is required when envSecret.enabled=true and envSecret.existingName is empty" -}}
{{- end -}}
{{- printf "%s-env-secret" (include "generic.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end }}

{{/*
Return the ConfigMap name used for file mounts.
*/}}
{{- define "generic.fileConfigName" -}}
{{- if .Values.fileConfig.existingName -}}
{{- .Values.fileConfig.existingName -}}
{{- else -}}
{{- if empty .Values.fileConfig.data -}}
{{- fail "fileConfig.data is required when fileConfig.enabled=true and fileConfig.existingName is empty" -}}
{{- end -}}
{{- printf "%s-file-config" (include "generic.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end }}

{{/*
Return the Secret name used for file mounts.
*/}}
{{- define "generic.fileSecretName" -}}
{{- if .Values.fileSecret.existingName -}}
{{- .Values.fileSecret.existingName -}}
{{- else -}}
{{- if empty .Values.fileSecret.stringData -}}
{{- fail "fileSecret.stringData is required when fileSecret.enabled=true and fileSecret.existingName is empty" -}}
{{- end -}}
{{- printf "%s-file-secret" (include "generic.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end }}

{{/*
Return a checksum for chart-managed ConfigMaps and Secrets.
Existing external objects are intentionally excluded because Helm cannot see their contents.
*/}}
{{- define "generic.configChecksum" -}}
{{- $payload := dict -}}
{{- if and .Values.envConfig.enabled (not .Values.envConfig.existingName) -}}
{{- $_ := set $payload "envConfig" .Values.envConfig.data -}}
{{- end -}}
{{- if and .Values.fileConfig.enabled (not .Values.fileConfig.existingName) -}}
{{- $_ := set $payload "fileConfig" .Values.fileConfig.data -}}
{{- end -}}
{{- if and .Values.envSecret.enabled (not .Values.envSecret.existingName) -}}
{{- $_ := set $payload "envSecret" .Values.envSecret.stringData -}}
{{- end -}}
{{- if and .Values.fileSecret.enabled (not .Values.fileSecret.existingName) -}}
{{- $_ := set $payload "fileSecret" .Values.fileSecret.stringData -}}
{{- end -}}
{{- toYaml $payload | sha256sum -}}
{{- end }}
