{{/*
Expand the name of the chart.
*/}}
{{- define "nvidia-gpu-list-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nvidia-gpu-list-exporter.fullname" -}}
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
{{- define "nvidia-gpu-list-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "nvidia-gpu-list-exporter.labels" -}}
helm.sh/chart: {{ include "nvidia-gpu-list-exporter.chart" . }}
{{ include "nvidia-gpu-list-exporter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/component: exporter
app.kubernetes.io/part-of: prometheus-monitoring
{{- end }}

{{/*
Selector labels
*/}}
{{- define "nvidia-gpu-list-exporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "nvidia-gpu-list-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "nvidia-gpu-list-exporter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "nvidia-gpu-list-exporter.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Allow the release namespace to be overridden for multi-namespace deployments in combined charts
*/}}
{{- define "nvidia-gpu-list-exporter.namespace" -}}
{{- if .Values.namespaceOverride }}
{{- .Values.namespaceOverride }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Create container image reference
*/}}
{{- define "nvidia-gpu-list-exporter.image" -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) }}
{{- end }}

{{/*
Generate environment variables for the exporter
*/}}
{{- define "nvidia-gpu-list-exporter.env" -}}
- name: EXPORTER_HOST
  value: {{ .Values.host | quote }}
- name: EXPORTER_PORT
  value: {{ .Values.port | quote }}
- name: EXPORTER_INTERVAL
  value: {{ .Values.exporter.interval | quote }}
- name: EXPORTER_TIMEOUT
  value: {{ .Values.exporter.timeout | quote }}
- name: NODE_NAME
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
{{- if .Values.exporter.logLevel }}
- name: LOG_LEVEL
  value: {{ .Values.exporter.logLevel | quote }}
{{- end }}
{{- if .Values.exporter.nvidiaSmiPath }}
- name: NVIDIA_SMI_PATH
  value: {{ .Values.exporter.nvidiaSmiPath | quote }}
{{- end }}
{{- if .Values.exporter.hostnameOverride }}
- name: HOSTNAME_OVERRIDE
  value: {{ .Values.exporter.hostnameOverride | quote }}
{{- end }}
{{- with .Values.nvidiaRuntime.visibleDevices }}
- name: NVIDIA_VISIBLE_DEVICES
  value: {{ . | quote }}
{{- end }}
{{- with .Values.nvidiaRuntime.driverCapabilities }}
- name: NVIDIA_DRIVER_CAPABILITIES
  value: {{ . | quote }}
{{- end }}
{{- range .Values.env }}
- name: {{ .name }}
  value: {{ .value | quote }}
{{- end }}
{{- end }}
