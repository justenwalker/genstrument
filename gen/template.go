package main

import (
	"embed"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/template"
)

//go:embed templates
var templatesFS embed.FS

var funcMap = map[string]interface{}{
	"quote": strconv.Quote,
	"arg_list": func(wf TemplateFunctionConfig) string {
		arglist := make([]string, 0, len(wf.Arguments))
		for _, a := range wf.Arguments {
			arglist = append(arglist, fmt.Sprintf("%s %s", a.Name, a.Type))
		}
		return strings.Join(arglist, ",")
	},
	"return_list": func(wf TemplateFunctionConfig) string {
		if len(wf.Returns) == 0 {
			return ""
		}
		arglist := make([]string, 0, len(wf.Returns))
		for _, r := range wf.Returns {
			arglist = append(arglist, fmt.Sprintf("%s %s", r.Name, r.Type))
		}
		return "(" + strings.Join(arglist, ",") + ")"
	},
	"call_list": func(wf TemplateFunctionConfig) string {
		arglist := make([]string, 0, len(wf.Arguments))
		for _, a := range wf.Arguments {
			arglist = append(arglist, a.Name)
		}
		return strings.Join(arglist, ",")
	},
	"assign_result_list": func(wf TemplateFunctionConfig) string {
		if len(wf.Returns) == 0 {
			return ""
		}
		arglist := make([]string, 0, len(wf.Returns))
		for _, r := range wf.Returns {
			arglist = append(arglist, r.Name)
		}
		return strings.Join(arglist, ",") + " = "
	},
}

func generateOutput(exp TemplateData, w io.Writer) error {
	t, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*")
	if err != nil {
		return err
	}
	if err = t.ExecuteTemplate(w, "template.tmpl", exp); err != nil {
		return fmt.Errorf("template execute: %w", err)
	}
	return nil
}
