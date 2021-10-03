{{/*
image to video fullname
*/}}
{{- define "slidesToVideo.imageToVideoFullname" -}}
{{ include "slidesToVideo.fullname" . }}-image-to-video
{{- end }}

{{/*
image to video common labels
*/}}
{{- define "slidesToVideo.imageToVideoLabels" -}}
{{ include "slidesToVideo.labels" . }}
app.kubernetes.io/component: image-to-video
{{- end }}

{{/*
image to video selector labels
*/}}
{{- define "slidesToVideo.imageToVideoSelectorLabels" -}}
{{ include "slidesToVideo.selectorLabels" . }}
app.kubernetes.io/component: image-to-video
{{- end }}

{{/*
image to video image
*/}}
{{- define "slidesToVideo.imageToVideoImage" -}}
{{- $dict := dict "service" .Values.imageToVideo.image "global" .Values.global.image "defaultVersion" .Chart.AppVersion -}}
{{- include "slidesToVideo.image" $dict -}}
{{- end }}