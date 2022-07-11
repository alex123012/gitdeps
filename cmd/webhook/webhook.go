package webhook

import (
	"github.com/spf13/cobra"
)

var fromFile bool

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "",
	}
	cmd.AddCommand(
		NewWebhookGenCmd(),
		NewWebHookCmd(),
		NewWebhookRemCmd(),
	)
	cmd.Flags().BoolVar(&fromFile, "from-file", false, "")
	return cmd
}
