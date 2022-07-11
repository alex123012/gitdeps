package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	dconfig "github.com/alex123012/gitdeps/pkg/config"
	v1 "k8s.io/api/admissionregistration/v1"

	"github.com/alex123012/gitdeps/cmd/check"
	"github.com/alex123012/gitdeps/cmd/common"
	"github.com/alex123012/gitdeps/cmd/config"
	"github.com/alex123012/gitdeps/cmd/version"
	"github.com/alex123012/gitdeps/cmd/webhook"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	logLevel string // "debug", "info", "warn", "error", "off"
	verbose  bool

	rootCmd = &cobra.Command{
		Use:           dconfig.ApplicationName,
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				logLevel = "debug"
			}

			if err := setLogLevel(logLevel); err != nil {
				return err
			}

			if err := common.Init(); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func Execute(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		hclog.L().Error(err.Error())
		os.Exit(1)
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	Execute(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"Path to the configuration file. (default \"$HOME/.config/gitdeps/config.yaml\")")

	rootCmd.PersistentFlags().StringVar(&logLevel, "log_level", "info",
		"Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, off]")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	rootCmd.AddCommand(
		config.NewCmd(),
		webhook.NewCmd(),
		check.NewCmd(),
		version.NewCmd(),
	)
}

func initConfig() {
	var err error
	if cfgFile == "" {
		cfgFile, err = dconfig.DefaultConfigPath()
		cobra.CheckErr(err)
	}

	_, err = os.Stat(cfgFile)
	if err != nil {
		hclog.L().Warn(fmt.Sprintf("can't read config: %s", err))
	}
	viper.SetConfigFile(cfgFile)

	rule := v1.RuleWithOperations{
		Operations: []v1.OperationType{"UPDATE", "CREATE", "DELETE"},
		Rule: v1.Rule{
			APIGroups:   []string{"apps", "networking.k8s.io", "extensions", ""},
			APIVersions: []string{"*"},
			Resources:   []string{"deployments", "ingresses", "statefulsets", "daemonsets", "services"},
		},
	}
	viper.SetDefault("webhook_conf.webhook.clientConfig.service.path", "/validate")
	viper.SetDefault("webhook_conf.webhook.clientConfig.service.port", 443)
	viper.SetDefault("webhook_conf.webhook.rules", rule)
	viper.SetDefault("webhook_conf.webhook.metadata.namespace", "validation-webhook")
	viper.SetDefault("webhook_conf.webhook.rules", rule)
	viper.SetDefault("webhook_conf.webhook.sideEffects", "None")
	viper.SetDefault("webhook_conf.webhook.admissionReviewVersions", "v1")

	path, err := filepath.Abs("./")
	if err != nil {
		hclog.L().Warn(fmt.Sprintf("Can't use abs path for: %s", path))
	}

	viper.SetDefault("git.target_branch", "")
	viper.SetDefault("git.compare_branch", "main")
	viper.SetDefault("git.path", path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		hclog.L().Debug("Using config file", "config", viper.ConfigFileUsed())
		return
	}
}

func setLogLevel(logLevel string) error {
	options := hclog.LoggerOptions{
		Level:             hclog.LevelFromString(logLevel),
		JSONFormat:        false,
		IncludeLocation:   false,
		DisableTime:       true,
		Color:             hclog.AutoColor,
		IndependentLevels: false,
	}

	hclog.SetDefault(hclog.New(&options))

	return nil
}
