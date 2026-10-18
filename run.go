package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type runResult struct {
	createdAt   time.Time
	completedAt time.Time
	output      string
	err         error
}

func runSteps(ctx context.Context, cfg *Config, stdout io.Writer, prog *progressPrinter) error {
	execID := newV7()
	prog.Printf("実行 %s", execID)

	var prevFiles, prevResponses []string
	total := len(cfg.Steps)
	for i, step := range cfg.Steps {
		last := i == total-1
		stepID := newV7()
		prog.Printf("[step %d/%d] 開始", i+1, total)

		text, err := buildPrompt(cfg, step, prevFiles, prevResponses)
		if err != nil {
			return fmt.Errorf(errStepFailed, i+1, err)
		}

		results := runParallel(ctx, cfg, i, total, step, !last, text, prog)

		var errs []error
		for j, r := range results {
			if r.err != nil {
				errs = append(errs, fmt.Errorf(errRunFailed, i+1, j+1, r.err))
			}
		}

		if last {
			if err := errors.Join(errs...); err != nil {
				return err
			}
			var sb strings.Builder
			for _, r := range results {
				sb.WriteString(r.output)
			}
			if err := writeOutput(cfg.Path(cfg.Output), stdout, sb.String()); err != nil {
				return err
			}
			prog.Printf("[step %d/%d] 完了", i+1, total)
			break
		}

		prevFiles, prevResponses = nil, nil
		dir := filepath.Join(cfg.TempdirAbs(), execID, stepID)
		for j, r := range results {
			if r.err != nil {
				continue
			}
			path := filepath.Join(dir, newV7()+".json")
			args := step.Run[j].Args
			if args == nil {
				args = []string{}
			}
			err := writeResult(path, Result{
				Order:       j,
				CreatedAt:   r.createdAt,
				CompletedAt: r.completedAt,
				Run:         RunSpec{Type: step.Run[j].Type, Args: args},
				Response:    r.output,
			})
			if err != nil {
				errs = append(errs, err)
			}
			prevFiles = append(prevFiles, path)
			prevResponses = append(prevResponses, r.output)
		}
		if err := errors.Join(errs...); err != nil {
			return err
		}
		prog.Printf("[step %d/%d] 完了 %s", i+1, total, dir)
	}
	return nil
}

func buildPrompt(cfg *Config, step Step, prevFiles, prevResponses []string) (string, error) {
	if len(step.Input) == 0 {
		return concatFiles(prevFiles)
	}
	d, err := newPromptData(cfg.Env, prevResponses)
	if err != nil {
		return "", err
	}
	paths := make([]string, len(step.Input))
	for i, p := range step.Input {
		paths[i] = cfg.Path(p)
	}
	return expandFiles(paths, d)
}

func runParallel(ctx context.Context, cfg *Config, stepIdx, total int, step Step, jsonOutput bool, text string, prog *progressPrinter) []runResult {
	results := make([]runResult, len(step.Run))
	var wg sync.WaitGroup
	for j, r := range step.Run {
		wg.Add(1)
		go func(j int, r Run) {
			defer wg.Done()
			label := fmt.Sprintf("[step %d/%d] run %d %s", stepIdx+1, total, j+1, r.Type)

			rctx := ctx
			if cfg.TimeoutDuration > 0 {
				var cancel context.CancelFunc
				rctx, cancel = context.WithTimeout(ctx, cfg.TimeoutDuration)
				defer cancel()
			}

			prog.Printf("%s 開始", label)
			res := runResult{createdAt: time.Now().UTC()}
			res.output, res.err = execCLI(rctx, r.Type, cfg.WorkdirAbs, r.Args, jsonOutput, text)
			res.completedAt = time.Now().UTC()
			elapsed := res.completedAt.Sub(res.createdAt).Round(time.Millisecond)
			if res.err != nil {
				prog.Printf("%s 失敗 (%s): %v", label, elapsed, res.err)
			} else {
				prog.Printf("%s 完了 (%s)", label, elapsed)
			}
			results[j] = res
		}(j, r)
	}
	wg.Wait()
	return results
}
