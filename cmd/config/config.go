package config

import (
	"fmt"

	"github.com/alex123012/gitdeps/cmd/common"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "",
	}

	cmd.AddCommand(
		NewRenderCmd(),
	)

	return cmd
}

func NewRenderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render",
		Short: "render config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return PrintStruct(common.Config)
		},
	}

	return cmd
}

func PrintStruct(structure interface{}) error {
	res, err := yaml.Marshal(structure)
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
}
