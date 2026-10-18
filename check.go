package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func checkConfig(c *Config) error {
	var errs []error
	addf := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf(format, args...))
	}

	if len(c.Steps) == 0 {
		addf(errStepsRequired)
	}

	for i, s := range c.Steps {
		at := fmt.Sprintf("steps[%d]", i)

		if len(s.Run) == 0 {
			addf(errRunRequired, at)
		}
		for j, r := range s.Run {
			if r.Type == "" {
				addf(errTypeRequired, at, j)
			} else if _, _, ok := command(r.Type, nil, false); !ok {
				addf(errTypeUnsupported, at, j, r.Type, strings.Join(cliTypes, ", "))
			}
		}

		if i == 0 && len(s.Input) == 0 {
			addf(errFirstInputRequired, at)
		}

		for j, p := range s.Input {
			where := fmt.Sprintf("%s.input[%d] (%s)", at, j, p)
			b, err := os.ReadFile(c.Path(p))
			if err != nil {
				addf("%s: %v", where, err)
				continue
			}
			t, err := parseTemplate(p, string(b))
			if err != nil {
				addf("%s: %v", where, err)
				continue
			}
			for _, e := range checkRefs(c, i, collectRefs(t)) {
				addf("%s: %s", where, e)
			}
		}
	}
	return errors.Join(errs...)
}

func checkRefs(c *Config, step int, refs []templateRef) []string {
	var errs []string
	for _, r := range refs {
		switch r.Root {
		case "Input", "Inputs", "Response", "Responses":
			if step == 0 {
				errs = append(errs, fmt.Sprintf(errRefInFirstStep, r.Root))
				continue
			}
			n := len(c.Steps[step-1].Run)
			if (r.Root == "Input" || r.Root == "Response") && n > 1 {
				errs = append(errs, fmt.Sprintf(errRefSingleOfMultiple, r.Root, n, r.Root))
			}
		case "Env":
			if r.Field != "" {
				if _, ok := c.Env[r.Field]; !ok {
					errs = append(errs, fmt.Sprintf(errEnvUndefined, r.Field))
				}
			}
		default:
			errs = append(errs, fmt.Sprintf(errUnknownRef, r.Root))
		}
	}
	return errs
}
