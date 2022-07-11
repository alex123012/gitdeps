package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/alex123012/gitdeps/pkg/config"

	"github.com/hashicorp/go-hclog"
	"github.com/xanzy/go-gitlab"
)

type Client struct {
	Hosts Hosts

	config *config.Config
}

type Hosts struct {
	hosts  []*Host
	mapper map[string]*Host
}
type Host struct {
	Name, URL string
	Client    *gitlab.Client
}

func (h *Hosts) GetHost(url string) (*Host, bool) {
	url = TrimUrl(url)
	value, err := h.mapper[url]
	return value, err
}

func (h *Hosts) append(host config.Host, name string, gitlabClient *gitlab.Client) error {
	if h.mapper == nil {
		h.mapper = make(map[string]*Host)
	}

	value := &Host{
		Name:   name,
		URL:    host.URL,
		Client: gitlabClient,
	}
	h.hosts = append(h.hosts, value)
	h.mapper[TrimUrl(host.URL)] = value

	return nil
}

func NewClient(cfg *config.Config) (*Client, error) {

	var options []gitlab.ClientOptionFunc

	if hclog.L().IsDebug() {
		options = append(options, gitlab.WithCustomLeveledLogger(hclog.Default().Named("go-gitlab")))
	}

	client := Client{config: cfg}
	for name, host := range cfg.Hosts {
		if host.URL == "" {
			return nil, fmt.Errorf("missing url for host %q", name)
		}
		if host.Token == "" {
			return nil, fmt.Errorf("missing token for host %q", name)
		}
		if !host.RateLimiter.Enabled {
			options = append(options, gitlab.WithCustomLimiter(&FakeLimiter{}))
		}
		gl, err := gitlab.NewClient(host.Token,
			append(options, gitlab.WithBaseURL(host.URL))...)
		if err != nil {
			return nil, err
		}
		err = client.Hosts.append(host, name, gl)
		if err != nil {
			return nil, err
		}
	}

	return &client, nil

}

// Used to avoid unnecessary noncached requests
type FakeLimiter struct{}

func (*FakeLimiter) Wait(context.Context) error {
	return nil
}

func TrimUrl(url string) string {
	return strings.Trim(url, "htps:/")
}
