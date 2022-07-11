package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ApplicationName = "gitdeps"
)

type Config struct {
	Hosts       Hosts       `yaml:"hosts" mapstructure:"hosts"`
	WebhookConf WebHookConf `yaml:"webhook_conf" mapstructure:"webhook_conf"`
	Git         GitConfig   `yaml:"git" mapstructure:"git"`
}

type GitConfig struct {
	TargetBranch  string `yaml:"target_branch" mapstructure:"target_branch"`
	CompareBranch string `yaml:"compare_branch" mapstructure:"compare_branch"`
	Path          string `yaml:"path" mapstructure:"path"`
}
type WebHookConf struct {
	Metadata metav1.ObjectMeta    `yaml:"metadata" mapstructure:"metadata"`
	Webhook  v1.ValidatingWebhook `yaml:"webhook" mapstructure:"webhook"`
	Tls      CertificateConf      `yaml:"tls" mapstructure:"tls"`
}

type CertificateConf struct {
	KeyFile      string `yaml:"key_file" mapstructure:"key_file"`
	CertFile     string `yaml:"cert_file" mapstructure:"cert_file"`
	Organization string `yaml:"organization" mapstructure:"organization"`
	Path         string `yaml:"path" mapstructure:"path"`
}
type Hosts map[string]Host

type Host struct {
	URL         string             `yaml:"url" mapstructure:"url"`
	Token       string             `yaml:"token" mapstructure:"token"`
	RateLimiter RateLimiterOptions `yaml:"rate_limiter" mapstructure:"rate_limiter"`
}

type RateLimiterOptions struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
}

func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", ApplicationName, "config.yaml"), err
}

func FromFile(path string) (*Config, error) {
	var config Config
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %q error: %v", path, err)
	}

	return &config, nil
}
