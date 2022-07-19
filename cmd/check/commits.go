package check

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hashicorp/go-hclog"

	"github.com/go-git/go-git/v5/plumbing/revlist"
	"github.com/spf13/cobra"
)

var (
	compareRef = "origin/HEAD"
	targetRef  = "HEAD"
	path       = "./"
	fetch      = false
)

func NewDefaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commits",
		Short: "",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			haveAll, err := TargetHaveAllCommitsFromOtherBranch(ctx, fetch, path, compareRef, targetRef)
			if !haveAll && err == nil {
				return fmt.Errorf("branch '%s' haven't some commits from '%s'", targetRef, compareRef)
			}
			return err
		},
	}

	cmd.Flags().StringVar(&path, "git-path", path, "")

	cmd.Flags().StringVar(&compareRef, "compare-ref", compareRef, "")

	cmd.Flags().StringVar(&targetRef, "target-ref", targetRef, "")

	cmd.Flags().BoolVar(&fetch, "fetch", fetch, "")

	return cmd
}

func TargetHaveAllCommitsFromOtherBranch(ctx context.Context, fetch bool, path, compareRef, targetRef string) (bool, error) {

	r, err := git.PlainOpen(path)
	if err != nil {
		return false, err
	}
	if fetch {
		err := r.FetchContext(ctx, &git.FetchOptions{})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return false, err
		}
	}

	compareHash, err := ResolveRevisionBranchHead(r, compareRef)
	if err != nil {
		return false, err
	}

	targetHash, err := ResolveRevisionBranchHead(r, targetRef)
	if err != nil {
		return false, err
	}

	hclog.L().Info(fmt.Sprintf("HEAD ref: %s", targetHash))
	hclog.L().Info(fmt.Sprintf("Default ref: %s", compareHash))

	compareCommit, err := r.CommitObject(*compareHash)
	if err != nil {
		return false, err
	}

	targetCommit, err := r.CommitObject(*targetHash)
	if err != nil {
		return false, err
	}

	res, err := compareCommit.IsAncestor(targetCommit)
	if err != nil {
		return false, err
	}
	if hclog.L().IsDebug() {
		storer := r.Storer

		hashStore := make([]plumbing.Hash, 0)

		allObjectsIter, err := storer.IterEncodedObjects(plumbing.AnyObject)
		if err != nil {
			return false, err
		}

		allObjectsIter.ForEach(func(obj plumbing.EncodedObject) error {
			if obj.Type() != plumbing.CommitObject {
				hashStore = append(hashStore, obj.Hash())
			}
			return nil
		})

		commitHashList, err := revlist.Objects(storer, []plumbing.Hash{*compareHash}, append(hashStore, *targetHash))
		if err != nil {
			return false, err
		}
		commitHashListLen := len(commitHashList)
		if commitHashListLen > 0 {
			newLine := "\n"
			debugMsg := newLine + "[DEBUG] "
			var commitList []*object.Commit

			for _, commitHash := range commitHashList {
				commit, err := r.CommitObject(commitHash)
				if err != nil {
					return false, err
				}
				commitList = append(commitList, commit)
			}

			commitListSorted := CommitSlice(commitList)
			sort.Sort(commitListSorted)

			for _, commit := range commitListSorted {
				commitMsg := debugMsg + strings.ReplaceAll(commit.String(), newLine, debugMsg) + newLine

				hclog.L().Debug(fmt.Sprintf("Target (%s) ref doesn't have commit from %s:%s", targetRef, compareRef, commitMsg))
			}
		}
		if commitHashListLen > 0 && res {
			return false, fmt.Errorf(
				"error finding ancestor: %s is ancestor for %s, but there is no additional commits after %s",
				compareHash,
				targetHash,
				compareHash,
			)
		}

		if commitHashListLen == 0 && !res {
			return false, fmt.Errorf(
				"error finding ancestor: %s isn't ancestor for %s, but there is some additional commits after %s",
				compareHash,
				targetHash,
				compareHash,
			)
		}
	}

	if !res {
		return false, nil
	}
	return true, nil
}

func ResolveRevisionBranchHead(r *git.Repository, s string) (*plumbing.Hash, error) {
	return r.ResolveRevision(plumbing.Revision(s))
}

type CommitSlice []*object.Commit

func (h CommitSlice) Len() int           { return len(h) }
func (h CommitSlice) Less(i, j int) bool { return h[i].Author.When.Before(h[j].Author.When) }
func (h CommitSlice) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
