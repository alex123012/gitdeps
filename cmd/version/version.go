package version

import (
	"fmt"

	"github.com/alex123012/gitdeps/pkg/config"
	"github.com/flant/glaball/pkg/util"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(util.PrintVersion(config.ApplicationName))
		},
	}

	return cmd
}
