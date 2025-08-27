{{/* Annotations for stakater/Reloader to rolling-update upon changes in static resources (Secret, ConfigMap, ...) */}}
{{- define "reloader-annotations" }}
  annotations:
    reloader.stakater.com/auto: "true"
{{- end }}

