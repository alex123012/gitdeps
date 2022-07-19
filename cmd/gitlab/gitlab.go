package gitlab

import (
	"strconv"

	"github.com/alex123012/gitdeps/pkg/gitlab"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitlab",
		Short: "",
		RunE: func(_ *cobra.Command, _ []string) error {
			annot := "https://gitlab.easy7.ru/diginavis/diginavis-bros/pipelines/2742"
			dontHaveDefaultCommits, err := gitlab.TargetHaveAllCommitsFromDefault(annot)

			if err != nil {
				return err
			}
			hclog.L().Debug(strconv.FormatBool(dontHaveDefaultCommits))
			return nil
		},
	}

	return cmd
}
