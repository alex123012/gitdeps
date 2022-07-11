package webhook

import (
	"os"
	"path/filepath"

	"github.com/alex123012/gitdeps/cmd/common"
	"github.com/alex123012/gitdeps/pkg/webhook"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfigPath string
	local          bool
)

func NewWebhookGenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-webhook-configuration",
		Short: "Validating webhook",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()
			config, err := GenerateNewConfig(local)
			if err != nil {
				return err
			}

			caMap, err := webhook.GenerateCertificate(ctx, common.Config.WebhookConf, fromFile)
			if err != nil {
				return err
			}

			err = webhook.CreateWebhookConf(ctx, config, caMap["ca"], common.Config.WebhookConf)
			if err != nil {
				return err
			}

			return nil

		},
	}

	k8sCmdConfigFlags(cmd)
	return cmd
}

func NewWebhookRemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-webhook-configuration",
		Short: "Validating webhook",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()
			config, err := GenerateNewConfig(local)
			if err != nil {
				return err
			}

			api, err := webhook.AdmissionApiFromConfig(config)
			if err != nil {
				return err
			}
			err = api.Delete(ctx, common.Config.WebhookConf.Metadata.Name, metav1.DeleteOptions{})
			return err

		},
	}
	k8sCmdConfigFlags(cmd)
	return cmd
}

func k8sCmdConfigFlags(cmd *cobra.Command) {

	cmd.Flags().BoolVar(&local, "local", false, "Use local kubeconfig")

	cmd.Flags().StringVar(&kubeConfigPath, "kubeconfig",
		filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		"Path to kubeconfig to use when flag --local is true")
}

func GenerateNewConfig(local bool) (*rest.Config, error) {
	var clusterConfig *rest.Config
	var err error

	switch local {
	case true:
		clusterConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	case false:
		clusterConfig, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	return clusterConfig, nil
}
