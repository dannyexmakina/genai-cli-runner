package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fakeCLIs(t *testing.T, ollamaScript string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("シェルスクリプトを使うため")
	}
	bin := t.TempDir()
	claude := `#!/bin/sh
in=$(cat)
case "$*" in
  *--output-format\ json*) printf '{"joke":"claude:%s"}' "$*";;
  *) printf '%s' "$in";;
esac
`
	if ollamaScript == "" {
		ollamaScript = `#!/bin/sh
cat >/dev/null
printf '{"joke":"ollama {{ raw }}"}'
`
	}
	os.WriteFile(filepath.Join(bin, "claude"), []byte(claude), 0o755)
	os.WriteFile(filepath.Join(bin, "ollama"), []byte(ollamaScript), 0o755)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func setup(t *testing.T, yaml string, files map[string]string) *Config {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644)
	}
	path := filepath.Join(dir, "runner.yaml")
	os.WriteFile(path, []byte(yaml), 0o644)
	c, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	c.WorkdirAbs = dir
	if err := checkConfig(c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestRunExplicitInput(t *testing.T) {
	fakeCLIs(t, "")
	c := setup(t, `
tempdir: tmp
env: {ENV: dev}
steps:
  - input: [a.md, b.md]
    run: [{type: claude, args: [--model, m]}, {type: ollama, args: [llama3.2]}]
  - input: [merge.md]
    run: [{type: claude}]
`, map[string]string{
		"a.md":     "A{{ .Env.ENV }}",
		"b.md":     "B",
		"merge.md": "1={{ (index .Inputs 0).joke }} 2={{ (index .Inputs 1).joke }}",
	})

	var out bytes.Buffer
	if err := runSteps(context.Background(), c, &out, newProgress(io.Discard)); err != nil {
		t.Fatal(err)
	}
	want := "1=claude:-p --output-format json --model m 2=ollama {{ raw }}"
	if out.String() != want {
		t.Fatalf("got %q, want %q", out.String(), want)
	}

	files, _ := filepath.Glob(filepath.Join(c.WorkdirAbs, "tmp", "*", "*", "*.json"))
	if len(files) != 2 {
		t.Fatalf("中間ファイル数 %d", len(files))
	}
}

func TestRunImplicitInputAndOutputFile(t *testing.T) {
	fakeCLIs(t, "")
	c := setup(t, `
output: out/result.md
steps:
  - input: [a.md]
    run: [{type: claude}, {type: ollama}]
  - run: [{type: claude}, {type: claude}]
`, map[string]string{"a.md": "A"})

	if err := runSteps(context.Background(), c, io.Discard, newProgress(io.Discard)); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(c.WorkdirAbs, "out", "result.md"))
	if err != nil {
		t.Fatal(err)
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	var orders []int
	for dec.More() {
		var r struct {
			Order    int    `json:"order"`
			Response string `json:"response"`
		}
		if err := dec.Decode(&r); err != nil {
			t.Fatalf("%v\n%s", err, b)
		}
		orders = append(orders, r.Order)
	}
	if got := len(orders); got != 4 || orders[0] != 0 || orders[1] != 1 {
		t.Fatalf("orders=%v\n%s", orders, b)
	}
	if !strings.Contains(string(b), "{{ raw }}") {
		t.Fatalf("応答が展開されている:\n%s", b)
	}
}

func TestRunParallelFailureWaitsAll(t *testing.T) {
	fakeCLIs(t, "#!/bin/sh\ncat >/dev/null\nexit 3\n")
	c := setup(t, `
steps:
  - input: [a.md]
    run: [{type: ollama}, {type: claude}]
  - input: [a.md]
    run: [{type: claude}]
`, map[string]string{"a.md": "A"})

	var out bytes.Buffer
	err := runSteps(context.Background(), c, &out, newProgress(io.Discard))
	if err == nil || !strings.Contains(err.Error(), "step 1 run 1") {
		t.Fatalf("got %v", err)
	}
	if out.Len() != 0 {
		t.Fatal("後続のstepは実行しない")
	}
	files, _ := filepath.Glob(filepath.Join(c.WorkdirAbs, "*", "*", "*.json"))
	if len(files) != 1 {
		t.Fatalf("中間ファイル数 %d", len(files))
	}
}

func TestRunTimeout(t *testing.T) {
	fakeCLIs(t, "#!/bin/sh\ncat >/dev/null\nsleep 5\n")
	c := setup(t, `
timeout: 100ms
steps:
  - input: [a.md]
    run: [{type: ollama}]
`, map[string]string{"a.md": "A"})

	err := runSteps(context.Background(), c, io.Discard, newProgress(io.Discard))
	if err == nil || !strings.Contains(err.Error(), "タイムアウト") {
		t.Fatalf("got %v", err)
	}
}

func TestDryRun(t *testing.T) {
	c := setup(t, `
env: {ENV: dev}
steps:
  - input: [a.md]
    run: [{type: claude}]
    expected:
      - {type: string, key: joke, value: hello}
      - {key: id}
  - input: [b.md]
    run: [{type: claude}]
  - run: [{type: claude}]
`, map[string]string{
		"a.md": "{{ .Env.ENV }}",
		"b.md": "{{ .Input.joke }}|{{ .Response }}|{{ len .Input.id }}",
	})

	var out bytes.Buffer
	if err := dryRun(c, &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"dev", "hello|{@response}|16", "=== step 3/3 (claude) ===\n{@response}"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("%q を含むべき:\n%s", want, out.String())
		}
	}
}

func TestDryRunWithoutExpected(t *testing.T) {
	c := setup(t, `
steps:
  - input: [a.md]
    run: [{type: claude}]
  - input: [b.md]
    run: [{type: claude}]
`, map[string]string{"a.md": "A", "b.md": "{{ .Input.joke }}"})

	err := dryRun(c, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "expected がない") {
		t.Fatalf("got %v", err)
	}
}
