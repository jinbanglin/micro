package template

var (
	Plugin = `package main
{{if .Plugins}}
import ({{range .Plugins}}
	_ "github.com/jinbanglin/go-plugins/{{.}}"{{end}}
){{end}}
`
)
