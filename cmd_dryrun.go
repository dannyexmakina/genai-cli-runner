package main

import "github.com/spf13/cobra"

func newDryrunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dryrun [file]",
		Short: "プロンプトを展開して標準出力に流す",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadAndCheck(args)
			if err != nil {
				return err
			}
			return dryRun(cfg, cmd.OutOrStdout())
		},
	}
}
