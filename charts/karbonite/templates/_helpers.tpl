{{/*
Expand the name of the chart.
*/}}
{{- define "karbonite.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "karbonite.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "karbonite.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "karbonite.labels" -}}
helm.sh/chart: {{ include "karbonite.chart" . }}
{{ include "karbonite.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "karbonite.selectorLabels" -}}
app.kubernetes.io/name: {{ include "karbonite.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "karbonite.serviceAccountName" -}}
{{- default (include "karbonite.fullname" .) .Values.serviceAccount.name }}
{{- end }}

{{/*
Create the name of the controller clusterrole (and of its clusterrolebinding)
*/}}
{{- define "karbonite.controller.clusterrole" -}}
{{- printf "%s-controller" (include "karbonite.fullname" .) }}
{{- end }}

{{/*
Create the name of the leader election role (and of its rolebinding)
*/}}
{{- define "karbonite.leaderelection.role" -}}
{{- printf "%s-leader-election" (include "karbonite.fullname" .) }}
{{- end }}