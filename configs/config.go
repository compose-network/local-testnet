package configs

import (
	"errors"
	"fmt"
)

var Values Config

type (
	Config struct {
		L1            L1            `mapstructure:"l1"`
		L2            L2            `mapstructure:"l2"`
		Observability Observability `mapstructure:"observability"`
	}

	L1 struct {
	}

	L2 struct {
		L1ChainID             int                   `mapstructure:"l1-chain-id"`
		L1ElURL               string                `mapstructure:"l1-el-url"`
		L1ClURL               string                `mapstructure:"l1-cl-url"`
		Wallet                Wallet                `mapstructure:"wallet"`
		CoordinatorPrivateKey string                `mapstructure:"coordinator-private-key"`
		Repositories          map[string]Repository `mapstructure:"repositories"`
		ChainIDs              ChainIDs              `mapstructure:"chain-ids"`
		DeploymentTarget      string                `mapstructure:"deployment-target"`
		GenesisBalanceWei     string                `mapstructure:"genesis-balance-wei"`
	}

	ChainIDs struct {
		RollupA int `mapstructure:"rollup-a"`
		RollupB int `mapstructure:"rollup-b"`
	}

	Repository struct {
		URL    string `mapstructure:"url"`
		Branch string `mapstructure:"branch"`
	}

	Wallet struct {
		PrivateKey string `mapstructure:"private-key"`
		Address    string `mapstructure:"address"`
	}

	Observability struct {
	}
)

func (c *L2) Validate() error {
	var errs []error

	if c.L1ChainID == 0 {
		errs = append(errs, errors.New("l2.l1-chain-id is required"))
	}
	if c.L1ElURL == "" {
		errs = append(errs, errors.New("l2.l1-el-url is required"))
	}
	if c.L1ClURL == "" {
		errs = append(errs, errors.New("l2.l1-cl-url is required"))
	}
	if c.CoordinatorPrivateKey == "" {
		errs = append(errs, errors.New("l2.coordinator-private-key is required"))
	}
	if c.Wallet.PrivateKey == "" {
		errs = append(errs, errors.New("l2.wallet.private-key is required"))
	}
	if c.Wallet.Address == "" {
		errs = append(errs, errors.New("l2.wallet.address is required"))
	}

	requiredRepos := []string{"op-geth", "optimism", "publisher"}
	for _, name := range requiredRepos {
		repo, exists := c.Repositories[name]
		if !exists {
			errs = append(errs, fmt.Errorf("l2.repositories.%s is required", name))
			continue
		}
		if repo.URL == "" {
			errs = append(errs, fmt.Errorf("l2.repositories.%s.url is required", name))
		}
		if repo.Branch == "" {
			errs = append(errs, fmt.Errorf("l2.repositories.%s.branch is required", name))
		}
	}

	if c.ChainIDs.RollupA == 0 {
		errs = append(errs, errors.New("l2.chain-ids.rollup-a is required"))
	}
	if c.ChainIDs.RollupB == 0 {
		errs = append(errs, errors.New("l2.chain-ids.rollup-b is required"))
	}

	if c.DeploymentTarget == "" {
		errs = append(errs, errors.New("l2.deployment-target is required"))
	} else if c.DeploymentTarget != "live" && c.DeploymentTarget != "calldata" {
		errs = append(errs, errors.New("l2.deployment-target must be either 'live' or 'calldata'"))
	}

	if len(errs) > 0 {
		return fmt.Errorf("L2 configuration validation failed: %w", errors.Join(errs...))
	}

	return nil
}
