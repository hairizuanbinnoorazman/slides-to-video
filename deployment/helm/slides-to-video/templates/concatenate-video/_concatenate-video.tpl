{{/*
concatenate video fullname
*/}}
{{- define "slidesToVideo.concatenateVideoFullname" -}}
{{ include "slidesToVideo.fullname" . }}-concatenate-video
{{- end }}

{{/*
concatenate video common labels
*/}}
{{- define "slidesToVideo.concatenateVideoLabels" -}}
{{ include "slidesToVideo.labels" . }}
app.kubernetes.io/component: concatenate-video
{{- end }}

{{/*
concatenate video selector labels
*/}}
{{- define "slidesToVideo.concatenateVideoSelectorLabels" -}}
{{ include "slidesToVideo.selectorLabels" . }}
app.kubernetes.io/component: concatenate-video
{{- end }}

{{/*
concatenate video image
*/}}
{{- define "slidesToVideo.concatenateVideoImage" -}}
{{- $dict := dict "service" .Values.concatenateVideo.image "global" .Values.global.image "defaultVersion" .Chart.AppVersion -}}
{{- include "slidesToVideo.image" $dict -}}
{{- end }}