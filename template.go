package docs

var MarkdownDocTemplate = `{{if gt .SectionNum 0}}% {{ .Command.Name }} {{ .SectionNum }}

{{end}}# NAME

{{ .Command.Name }}{{ if .Command.Usage }} - {{ .Command.Usage }}{{ end }}

# SYNOPSIS

{{ .Command.Name }}
{{ if .SynopsisArgs }}
` + "```" + `
{{ range $v := .SynopsisArgs }}{{ $v }}{{ end }}` + "```" + `
{{ end }}{{ if .Command.Description }}
# DESCRIPTION

{{ .Command.Description }}
{{ end }}
**Usage**:

` + "```" + `{{ if .Command.UsageText }}
{{ .Command.UsageText }}
{{ else }}
{{ .Command.Name }} [GLOBAL OPTIONS] [command [COMMAND OPTIONS]] [ARGUMENTS...]
{{ end }}` + "```" + `
{{ if .GlobalArgs }}
# GLOBAL OPTIONS
{{ range $v := .GlobalArgs }}
{{ $v }}{{ end }}
{{ end }}{{ if .Commands }}
# COMMANDS
{{ range $v := .Commands }}
{{ $v }}{{ end }}{{ end }}`

var MarkdownTabularDocTemplate = `{{ define "flags" }}
| Name | Description | Default value | Environment variables |
|------|-------------|:-------------:|:---------------------:|
{{   range $flag := . -}}
{{- /**/ -}} | ` + "`" + `{{ $flag.Name }}{{ if $flag.TakesValue }}="…"{{ end }}` + "`" + ` {{ if $flag.Aliases }}(` + "`" + `{{ join $flag.Aliases "` + "`, `" + `" }}` + "`" + `) {{ end }}
{{- /**/ -}} | {{ $flag.Usage }}
{{- /**/ -}} | {{ if $flag.Default }}` + "`" + `{{ $flag.Default }}` + "`" + `{{ end }}
{{- /**/ -}} | {{ if $flag.EnvVars }}` + "`" + `{{ join $flag.EnvVars "` + "`, `" + `" }}` + "`" + `{{ else }}*none*{{ end }}
{{- /**/ -}} |
{{   end }}
{{ end }}

{{ define "command" }}
### ` + "`" + `{{ .Name }}` + "`" + ` {{ if gt .Level 0 }}sub{{ end }}command{{ if .Aliases }} (aliases: ` + "`" + `{{ join .Aliases "` + "`, `" + `" }}` + "`" + `){{ end }}
{{   if .Usage }}
{{     .Usage }}.
{{   end }}
{{   if .UsageText }}
{{     range $line := .UsageText -}}
> {{ $line }}
{{     end -}}
{{   end }}
{{   if .Description }}
{{     .Description }}.
{{   end }}
Usage:

` + "```" + `bash
$ {{ .AppPath }} [GLOBAL FLAGS] {{ .Name }}{{ if .Flags }} [COMMAND FLAGS]{{ end }} {{ if .ArgsUsage }}{{ .ArgsUsage }}{{ else }}[ARGUMENTS...]{{ end }}
` + "```" + `

{{   if .Flags -}}
The following flags are supported:
{{     template "flags" .Flags }}
{{   end -}}

{{   if .SubCommands -}}
{{     range $subCmd := .SubCommands -}}
{{       template "command" $subCmd }}
{{     end -}}
{{   end -}}
{{ end }}

## CLI interface{{ if .Name }} - {{ .Name }}{{ end }}

{{ if .Description }}{{ .Description }}.
{{ end }}
{{ if .Usage }}{{ .Usage }}.
{{ end }}
{{ if .UsageText }}
{{   range $line := .UsageText -}}
> {{ $line }}
{{   end -}}
{{ end }}
Usage:

` + "```" + `bash
$ {{ .AppPath }}{{ if .GlobalFlags }} [GLOBAL FLAGS]{{ end }} [COMMAND] [COMMAND FLAGS] {{ if .ArgsUsage }}{{ .ArgsUsage }}{{ else }}[ARGUMENTS...]{{ end }}
` + "```" + `

{{ if .GlobalFlags }}
Global flags:

{{ template "flags" .GlobalFlags }}

{{ end -}}
{{ if .Commands -}}
{{   range $cmd := .Commands -}}
{{     template "command" $cmd }}
{{   end }}
{{- end }}`
