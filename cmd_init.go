package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [file]",
		Short: "テンプレートをカレントディレクトリに書き出す",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := fileArg(args)
			if _, err := os.Stat(file); !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf(errFileExists, file)
			}
			return os.WriteFile(file, defaultTemplate, 0o644)
		},
	}
}
