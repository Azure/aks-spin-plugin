package config

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env/v9"
	"github.com/olivermking/spin-aks-plugin/pkg/azure"
	"github.com/olivermking/spin-aks-plugin/pkg/logger"
	"github.com/olivermking/spin-aks-plugin/pkg/prompt"
	"github.com/olivermking/spin-aks-plugin/pkg/usererror"
)

var (
	c    config
	opts *Opts
)

// def sets empty options to their defaults
func (o *Opts) def() {
	if o == nil {
		o = &Opts{}
	}

	if o.Path == "" {
		o.Path = "./aks-spin.toml"
	}
}

// Load loads any current aks spin configs from a file or the environment, with precedence towards env variables
func Load(o Opts) error {
	opts = &o
	opts.def()

	if _, err := toml.DecodeFile(opts.Path, &c); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("decoding aks spin config toml file %s: %w", opts.Path, err)
	}

	if err := env.ParseWithOptions(&c, env.Options{
		Prefix:                "AKS_SPIN_",
		UseFieldNameByDefault: true,
	}); err != nil {
		return fmt.Errorf("parsing aks spin config from env variables: %w", err)
	}

	return nil
}

func EnsureCluster(ctx context.Context) error {
	lgr := logger.FromContext(ctx)
	lgr.Debug("starting to ensure cluster config")

	if c.Cluster.Subscription == "" {
		subs, err := azure.ListSubscriptions(ctx)
		if err != nil {
			return fmt.Errorf("listing subscriptions: %w", err)
		}

		lgr.Debug("prompting for cluster subscription")
		sub, err := prompt.Select("Select your Cluster's Subscription", subs, &prompt.SelectOpt[armsubscription.Subscription]{
			Field: func(t armsubscription.Subscription) string {
				return *t.DisplayName
			},
		})
		if err != nil {
			return fmt.Errorf("selecting subscription: %w", err)
		}
		c.Cluster.Subscription = *sub.SubscriptionID
		lgr.Debug("finished prompting for cluster subscription")
	}

	if c.Cluster.ResourceGroup == "" {
		rgs, err := azure.ListResourceGroups(ctx, c.Cluster.Subscription)
		if err != nil {
			return fmt.Errorf("listing resource groups: %w", err)
		}

		lgr.Debug("prompting for cluster resource group")
		rg, err := prompt.Select("Select your Cluster's Resource Group", rgs, &prompt.SelectOpt[armresources.ResourceGroup]{
			Field: func(t armresources.ResourceGroup) string {
				return *t.Name
			},
		})
		if err != nil {
			return fmt.Errorf("selecting resource group: %w", err)
		}
		c.Cluster.ResourceGroup = *rg.Name

		lgr.Debug("finished prompting for cluster resource group")
	}

	lgr.Debug("done ensuring cluster config")
	return nil
}

func fieldEmptyErr(field string) error {
	return usererror.New(fmt.Errorf("validating aks spin config: %s invalid", field), fmt.Sprintf("AKS Spin Config invalid. Field %s is empty.", field))
}

// Write writes the current aks spin config to a file
func Write() error {
	// create directories if they don't exist
	dirs := path.Dir(opts.Path)
	if _, err := os.Stat(dirs); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("validating directories %s: %w", dirs, err)
		}

		if err := os.MkdirAll(dirs, os.ModeDir|0755); err != nil {
			return fmt.Errorf("making directories %s: %w", dirs, err)
		}
	}

	// open file handles creating the file if it doesn't exist
	f, err := os.OpenFile(opts.Path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("encoding aks spin config toml file %s: %w", opts.Path, err)
	}

	return nil
}
