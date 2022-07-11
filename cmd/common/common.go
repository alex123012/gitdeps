package common

import (
	"github.com/alex123012/gitdeps/pkg/client"
	"github.com/alex123012/gitdeps/pkg/config"

	"github.com/spf13/viper"
)

var (
	Config *config.Config
	Client *client.Client
)

func Init() (err error) {

	var cfg config.Config
	if err = viper.Unmarshal(&cfg); err != nil {
		return err
	}

	Config = &cfg

	if Client, err = client.NewClient(Config); err != nil {
		return err
	}
	return nil
}
