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
{{- printf "%s-file-secret" (include "generic.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end }}

{{/*
Return the Secret name used as fileConfig render input.
*/}}
{{- define "generic.fileConfigRenderSecretName" -}}
{{- .Values.fileConfig.render.secret.existingName -}}
{{- end }}

{{/*
Validate config and secret values before rendering resources.
*/}}
{{- define "generic.validateConfig" -}}
{{- if and .Values.envConfig.enabled (not .Values.envConfig.existingName) (empty .Values.envConfig.data) -}}
{{- fail "envConfig requires either data or existingName when envConfig.enabled=true" -}}
{{- end -}}
{{- if and .Values.envSecret.enabled (not .Values.envSecret.existingName) (empty .Values.envSecret.stringData) -}}
{{- fail "envSecret requires either stringData or existingName when envSecret.enabled=true" -}}
{{- end -}}
{{- if and .Values.fileConfig.render.enabled (not .Values.fileConfig.enabled) -}}
{{- fail "fileConfig.enabled must be true when fileConfig.render.enabled=true" -}}
{{- end -}}
{{- if and .Values.fileConfig.enabled (not .Values.fileConfig.existingName) (empty .Values.fileConfig.data) -}}
{{- if .Values.fileConfig.render.enabled -}}
{{- fail "fileConfig requires either data or existingName when fileConfig.enabled=true; when fileConfig.render.enabled=true, data/existingName is used as the template source" -}}
{{- else -}}
{{- fail "fileConfig requires either data or existingName when fileConfig.enabled=true" -}}
{{- end -}}
{{- end -}}
{{- if and .Values.fileConfig.render.enabled (not .Values.fileConfig.render.secret.existingName) -}}
{{- fail "fileConfig.render.secret.existingName is required when fileConfig.render.enabled=true" -}}
{{- end -}}
{{- if and .Values.fileConfig.render.enabled (not .Values.fileConfig.render.secret.envFrom) (not .Values.fileConfig.render.secret.mountPath) -}}
{{- fail "fileConfig.render.secret requires envFrom=true or mountPath when fileConfig.render.enabled=true" -}}
{{- end -}}
{{- if and .Values.fileSecret.enabled (not .Values.fileSecret.existingName) (empty .Values.fileSecret.stringData) -}}
{{- fail "fileSecret requires either stringData or existingName when fileSecret.enabled=true" -}}
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
