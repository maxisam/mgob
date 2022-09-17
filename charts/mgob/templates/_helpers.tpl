{{/*
    Return the proper image name
*/}}
{{- define "mgob.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global) }}
{{- end -}}

{{/*
    Return the proper image name
*/}}
{{- define "mongodb.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.mongodb.image) }}
{{- end -}}

{{/*
 Create the name of the service account to use
 */}}
{{- define "mgob.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "common.names.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}