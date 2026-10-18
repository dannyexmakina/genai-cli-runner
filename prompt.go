package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

type promptData struct {
	Input     any
	Inputs    []any
	Response  string
	Responses []string
	Env       map[string]string
}

func newPromptData(env map[string]string, responses []string) (promptData, error) {
	d := promptData{Env: env, Responses: responses}
	for i, r := range responses {
		var v any
		if err := json.Unmarshal([]byte(r), &v); err != nil {
			return promptData{}, fmt.Errorf(errResponseNotJSON, i, err)
		}
		d.Inputs = append(d.Inputs, v)
	}
	if len(responses) == 1 {
		d.Input = d.Inputs[0]
		d.Response = responses[0]
	}
	return d, nil
}

func parseTemplate(name, text string) (*template.Template, error) {
	return template.New(name).Option("missingkey=error").Parse(text)
}

func expandFiles(paths []string, d promptData) (string, error) {
	var sb strings.Builder
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		t, err := parseTemplate(p, string(b))
		if err != nil {
			return "", err
		}
		if err := t.Execute(&sb, d); err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

func concatFiles(paths []string) (string, error) {
	var sb strings.Builder
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		sb.Write(b)
	}
	return sb.String(), nil
}
