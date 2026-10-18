package main

import (
	_ "embed"
	"os"

	"github.com/spf13/cobra"
)

const defaultFile = "runner.yaml"

//go:embed templates/01.yaml
var defaultTemplate []byte

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          "genai-cli-runner",
		Short:        "YAMLで宣言した手順に従いLLM CLIを実行する",
		SilenceUsage: true,
	}
	root.AddCommand(
		newInitCmd(),
		newConfigcheckCmd(),
		newDryrunCmd(),
		newRunCmd(),
	)
	return root
}

func fileArg(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	return defaultFile
}

func loadAndCheck(args []string) (*Config, error) {
	cfg, err := loadConfig(fileArg(args))
	if err != nil {
		return nil, err
	}
	if err := checkConfig(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
