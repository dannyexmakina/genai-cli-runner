package main

import (
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [file]",
		Short: "本番実行する",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadAndCheck(args)
			if err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
			defer stop()
			return runSteps(ctx, cfg, cmd.OutOrStdout(), newProgress(cmd.ErrOrStderr()))
		},
	}
}
