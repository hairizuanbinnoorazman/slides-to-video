{{/*
manager fullname
*/}}
{{- define "slidesToVideo.managerFullname" -}}
{{ include "slidesToVideo.fullname" . }}-manager
{{- end }}

{{/*
manager common labels
*/}}
{{- define "slidesToVideo.managerLabels" -}}
{{ include "slidesToVideo.labels" . }}
app.kubernetes.io/component: manager
{{- end }}

{{/*
manager selector labels
*/}}
{{- define "slidesToVideo.managerSelectorLabels" -}}
{{ include "slidesToVideo.selectorLabels" . }}
app.kubernetes.io/component: manager
{{- end }}

{{/*
manager image
*/}}
{{- define "slidesToVideo.managerImage" -}}
{{- $dict := dict "service" .Values.manager.image "global" .Values.global.image "defaultVersion" .Chart.AppVersion -}}
{{- include "slidesToVideo.image" $dict -}}
{{- end }}