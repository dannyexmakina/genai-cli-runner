package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var cliTypes = []string{"claude", "ollama"}

func command(typ string, args []string, jsonOutput bool) (name string, argv []string, ok bool) {
	switch typ {
	case "claude":
		name, argv = claudeCommand(args, jsonOutput)
	case "ollama":
		name, argv = ollamaCommand(args, jsonOutput)
	default:
		return "", nil, false
	}
	return name, argv, true
}

func claudeCommand(args []string, jsonOutput bool) (string, []string) {
	argv := []string{"-p"}
	if jsonOutput {
		argv = append(argv, "--output-format", "json")
	}
	return "claude", append(argv, args...)
}

func ollamaCommand(args []string, jsonOutput bool) (string, []string) {
	argv := append([]string{"run"}, args...)
	if jsonOutput {
		argv = append(argv, "--format", "json")
	}
	return "ollama", argv
}

func execCLI(ctx context.Context, typ string, dir string, args []string, jsonOutput bool, prompt string) (string, error) {
	name, argv, ok := command(typ, args, jsonOutput)
	if !ok {
		return "", fmt.Errorf(errUnsupportedCLI, typ)
	}
	cmd := exec.CommandContext(ctx, name, argv...)
	cmd.Dir = dir
	cmd.WaitDelay = time.Second
	cmd.Stdin = strings.NewReader(prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf(errCLITimeout, name)
		}
		return "", fmt.Errorf(errCLIFailed, name, err, strings.TrimSpace(stderr.String()))
	}
	out := stdout.String()
	if jsonOutput && !json.Valid([]byte(out)) {
		return "", fmt.Errorf(errCLINotJSON, name)
	}
	return out, nil
}
