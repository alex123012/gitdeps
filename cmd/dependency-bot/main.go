package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alex123012/dependency-bot/pkg/common"
	"github.com/alex123012/dependency-bot/pkg/gitlab"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
)

func main() {
	var (
		debug            bool
		gitServer        string
		token            string
		url              string
		port             int
		defaultTimer     int
		defaultTimerRepo int
		defaultSleepApi  int
		retriesOnFailure int
		repositoriesIds  []int
		controller       common.ControllerInterface
		gitlabType       = "gitlab"
		githubType       = "github"
		gitServerSlice   = []string{gitlabType, githubType + " (in future)"}
	)

	run := func(ctx context.Context) error {
		if gitServer == gitlabType {
			controller = gitlab.New(
				token,
				url,
				defaultTimer,
				defaultTimerRepo,
				defaultSleepApi,
				retriesOnFailure,
				repositoriesIds,
			)
		} else if gitServer == githubType {
			return fmt.Errorf("github support will be in the future versions")
		} else {
			return fmt.Errorf("not valid git-server flag value, use: %v", gitServerSlice)
		}
		klog.Infof("Starting controller %s", controller.GetName())
		grp, ctx := errgroup.WithContext(ctx)
		grp.Go(func() error {
			return controller.Run(ctx)
		})
		if debug {
			grp.Go(func() error {
				return Debug(ctx, port)
			})
		}
		return grp.Wait()
	}
	cmd := &cobra.Command{
		Use:     "dependency-bot",
		Short:   "Manage dependencies from default branch of repository to other branches",
		Version: "0.0.1",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			if err := run(ctx); err != nil {
				klog.Exitln(err)
			}
		},
	}

	flags := cmd.PersistentFlags()
	flags.BoolVar(&debug, "debug", false, "debug mode (default: false)")
	flags.StringVar(&gitServer, "git-server", gitlabType, fmt.Sprintf("The git server name\nsupported servers: %v", gitServerSlice))
	flags.StringVarP(&token, "token", "t", os.Getenv("DEPENDECIES_TOKEN"), "Token for git server api (default $DEPENDECIES_TOKEN)")
	flags.StringVar(&url, "url", "gitlab.com", "Url of Git server")
	flags.IntVarP(&port, "port", "p", 8888, "Port to use for debug server")
	flags.IntVar(&defaultTimer, "timer", 10, "Specify the time in which each repo will be checked, seconds")
	flags.IntVarP(&defaultTimerRepo, "timer-repo", "T", 24, "Specify the time in which repositories list will be updated, hours")
	flags.IntVar(&defaultSleepApi, "api-wait", 300, "Specify the the time in which bot will make one api request, miliseconds")
	flags.IntVar(&retriesOnFailure, "retries", 3, "Specify how many retries would bot make for one api request if request fails")
	flags.IntSliceVarP(&repositoriesIds, "repositories-ids", "R", []int{}, "Specifies the repositories ids to watch")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := cmd.ExecuteContext(ctx); err != nil {
		klog.Exitln(err)
	}
}
