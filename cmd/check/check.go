package check

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "",
	}

	cmd.AddCommand(
		NewDefaultCmd(),
	)

	return cmd
}
