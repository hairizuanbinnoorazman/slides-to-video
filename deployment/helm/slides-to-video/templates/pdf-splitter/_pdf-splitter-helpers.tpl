{{/*
pdf splitter fullname
*/}}
{{- define "slidesToVideo.pdfSplitterFullname" -}}
{{ include "slidesToVideo.fullname" . }}-pdf-splitter
{{- end }}

{{/*
pdf splitter common labels
*/}}
{{- define "slidesToVideo.pdfSplitterLabels" -}}
{{ include "slidesToVideo.labels" . }}
app.kubernetes.io/component: pdf-splitter
{{- end }}

{{/*
pdf splitter selector labels
*/}}
{{- define "slidesToVideo.pdfSplitterSelectorLabels" -}}
{{ include "slidesToVideo.selectorLabels" . }}
app.kubernetes.io/component: pdf-splitter
{{- end }}

{{/*
pdf splitter image
*/}}
{{- define "slidesToVideo.pdfSplitterImage" -}}
{{- $dict := dict "service" .Values.pdfSplitter.image "global" .Values.global.image "defaultVersion" .Chart.AppVersion -}}
{{- include "slidesToVideo.image" $dict -}}
{{- end }}