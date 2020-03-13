{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "cluster-operator.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cluster-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "cluster-operator.labels" -}}
helm.sh/chart: {{ include "cluster-operator.chart" . }}
{{ include "cluster-operator.selectorLabels" . }}
app.giantswarm.io/branch: {{ .Values.project.branch }}
app.giantswarm.io/commit: {{ .Values.project.commit }}
app.kubernetes.io/name: {{ include "cluster-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- include "cluster-operator.name" . }}.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "cluster-operator.selectorLabels" -}}
app: {{ include "cluster-operator.name" . }}
version: {{ .Chart.Version }}
{{- end -}}
