package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigcheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configcheck [file]",
		Short: "YAMLを静的検証する",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := loadAndCheck(args); err != nil {
				return err
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "ok")
			return nil
		},
	}
}
