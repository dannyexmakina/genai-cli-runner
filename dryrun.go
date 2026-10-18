package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

const placeholderResponse = "{@response}"

func dryRun(cfg *Config, w io.Writer) error {
	total := len(cfg.Steps)
	for i, step := range cfg.Steps {
		text, err := dryrunPrompt(cfg, i)
		if err != nil {
			return fmt.Errorf(errStepFailed, i+1, err)
		}
		types := make([]string, len(step.Run))
		for j, r := range step.Run {
			types[j] = r.Type
		}
		fmt.Fprintf(w, "=== step %d/%d (%s) ===\n%s\n", i+1, total, strings.Join(types, ", "), text)
	}
	return nil
}

func dryrunPrompt(cfg *Config, i int) (string, error) {
	step := cfg.Steps[i]
	d := promptData{Env: cfg.Env}
	var n int
	var prev Step
	if i > 0 {
		prev = cfg.Steps[i-1]
		n = len(prev.Run)
	}

	if len(step.Input) == 0 {
		return strings.Repeat(placeholderResponse, n), nil
	}

	if i > 0 {
		value := expectedValue(prev.Expected)
		for j := 0; j < n; j++ {
			d.Inputs = append(d.Inputs, value)
			d.Responses = append(d.Responses, placeholderResponse)
		}
		if n == 1 {
			d.Input = d.Inputs[0]
			d.Response = placeholderResponse
		}
	}

	paths := make([]string, len(step.Input))
	for j, p := range step.Input {
		paths[j] = cfg.Path(p)
	}
	text, err := expandFiles(paths, d)
	if err != nil && i > 0 && len(prev.Expected) == 0 {
		return "", fmt.Errorf(errNoExpected, err)
	}
	return text, err
}

func expectedValue(exp []Expected) map[string]any {
	m := map[string]any{}
	for _, e := range exp {
		if e.Type == "" || e.Value == nil {
			m[e.Key] = randomString()
			continue
		}
		m[e.Key] = e.Value
	}
	return m
}

func randomString() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
