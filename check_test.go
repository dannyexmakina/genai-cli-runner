package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func load(t *testing.T, dir, yaml string) *Config {
	t.Helper()
	path := filepath.Join(dir, "runner.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	c.WorkdirAbs = dir
	return c
}

func TestCheckOK(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"first.md":  "{{ .Env.ENV }}",
		"second.md": "{{ (index .Inputs 1).joke }} {{ index .Responses 0 }} {{ range .Inputs }}{{ .joke }}{{ end }}",
		"third.md":  "{{ .Input.joke }} {{ .Response }} {{ $.Env.ENV }}",
	})
	c := load(t, dir, `
env: {ENV: dev}
steps:
  - input: [first.md]
    run: [{type: claude}, {type: ollama}]
  - input: [second.md]
    run: [{type: claude}]
  - input: [third.md]
    run: [{type: claude}]
`)
	if err := checkConfig(c); err != nil {
		t.Fatal(err)
	}
}

func TestCheckErrors(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"first.md":  "{{ .Inputs }} {{ .Env.NOPE }} {{ .Foo }}",
		"second.md": "{{ .Input.joke }}",
		"broken.md": "{{ .Env",
	})
	c := load(t, dir, `
steps:
  - input: [first.md, missing.md, broken.md]
    run: [{type: claude}, {type: gpt}]
  - input: [second.md]
    run: [{}]
  - run: []
`)
	err := checkConfig(c)
	if err == nil {
		t.Fatal("エラーになるべき")
	}
	for _, want := range []string{
		`"gpt" は未対応`,
		".Inputs: 最初のstep",
		".Env.NOPE: env で定義されていない",
		".Foo: 未知の参照",
		"missing.md",
		"broken.md",
		"steps[1].run[0].type: 必須",
		".Input: 前ステップの出力が2個",
		"steps[2].run: 1つ以上必要",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%q を含むべき:\n%v", want, err)
		}
	}
}

func TestCheckFirstStepNeedsInput(t *testing.T) {
	c := load(t, t.TempDir(), `
steps:
  - run: [{type: claude}]
`)
	if err := checkConfig(c); err == nil || !strings.Contains(err.Error(), "最初のstepでは必須") {
		t.Fatalf("got %v", err)
	}
}

func TestLoadRejectsUnknownKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runner.yaml")
	os.WriteFile(path, []byte("stpes: []\n"), 0o644)
	if _, err := loadConfig(path); err == nil {
		t.Fatal("未知のキーはエラーになるべき")
	}
}

func TestLoadExpandsEnv(t *testing.T) {
	t.Setenv("GCR_TEST", "value")
	dir := t.TempDir()
	path := filepath.Join(dir, "runner.yaml")
	os.WriteFile(path, []byte("env: {A: \"${GCR_TEST}\"}\ntimeout: 3m\n"), 0o644)
	c, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.Env["A"] != "value" || c.TimeoutDuration.Minutes() != 3 {
		t.Fatalf("got env=%v timeout=%v", c.Env, c.TimeoutDuration)
	}
}
