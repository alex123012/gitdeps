package check

import (
	"context"
	"errors"
	"fmt"

	"github.com/alex123012/gitdeps/cmd/common"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
)

var (
	path, compareBranch, targetBranch string
)

func NewDefaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commits",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if path == "" {
				path = common.Config.Git.Path
			}

			if compareBranch == "" {
				compareBranch = common.Config.Git.CompareBranch
			}

			if targetBranch == "" {
				targetBranch = common.Config.Git.TargetBranch
			}

			haveAll, err := TargetHaveAllCommitsFromOtherBranch(ctx, path, compareBranch, targetBranch)
			if !haveAll && err == nil {
				return fmt.Errorf("branch '%s' haven't some commits from '%s'", targetBranch, compareBranch)
			}
			return err
		},
	}

	cmd.Flags().StringVar(&path, "git-path", "", "")

	cmd.Flags().StringVar(&compareBranch, "compare-branch", "", "")

	cmd.Flags().StringVar(&targetBranch, "target-branch", "", "default current HEAD")

	return cmd
}

func GetBranchRef(r *git.Repository, branchName string) (*plumbing.Reference, error) {

	iterator, err := r.Branches()
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var branchNameRef *plumbing.Reference
	iterator.ForEach(func(t *plumbing.Reference) error {
		if branchName == t.Name().Short() {
			branchNameRef = t
		}
		return nil
	})
	if branchNameRef == nil {
		return nil, fmt.Errorf("can't find branch with name '%s'", branchName)
	}
	return branchNameRef, nil
}

func TargetHaveAllCommitsFromOtherBranch(ctx context.Context, path, compareBranch, targetBranch string) (bool, error) {

	r, err := git.PlainOpen(path)
	if err != nil {
		return false, err
	}
	err = r.FetchContext(ctx, &git.FetchOptions{})
	if !(err == nil || errors.Is(err, git.NoErrAlreadyUpToDate)) {
		return false, err
	}

	var headRef *plumbing.Reference
	if targetBranch == "" {
		headRef, err = r.Head()
	} else {
		headRef, err = GetBranchRef(r, targetBranch)
	}

	if err != nil {
		return false, err
	}

	compareBranchRef, err := GetBranchRef(r, compareBranch)
	if err != nil {
		return false, err
	}
	commitHead, err := r.CommitObject(headRef.Hash())
	if err != nil {
		return false, err
	}
	commitCompare, err := r.CommitObject(compareBranchRef.Hash())
	if err != nil {
		return false, err
	}
	patch, err := commitHead.Patch(commitCompare)
	if err != nil {
		return false, err
	}

	diff := patch.FilePatches()
	if len(diff) > 0 {
		return false, nil
	}
	return true, nil
}
