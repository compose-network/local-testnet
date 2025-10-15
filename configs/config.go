package configs

import (
	"errors"
	"fmt"
)

var Values Config

type (
	RepositoryName string
	L2ChainName    string

	Config struct {
		L1            L1            `mapstructure:"l1"`
		L2            L2            `mapstructure:"l2"`
		Observability Observability `mapstructure:"observability"`
	}

	L1 struct {
	}

	L2 struct {
		L1ChainID             int                           `mapstructure:"l1-chain-id"`
		L1ElURL               string                        `mapstructure:"l1-el-url"`
		L1ClURL               string                        `mapstructure:"l1-cl-url"`
		Wallet                Wallet                        `mapstructure:"wallet"`
		CoordinatorPrivateKey string                        `mapstructure:"coordinator-private-key"`
		Repositories          map[RepositoryName]Repository `mapstructure:"repositories"`
		ChainConfigs          map[L2ChainName]ChainConfig   `mapstructure:"chain-configs"`
		DeploymentTarget      string                        `mapstructure:"deployment-target"`
		GenesisBalanceWei     string                        `mapstructure:"genesis-balance-wei"`
	}

	ChainConfig struct {
		ID      int `mapstructure:"id"`
		RPCPort int `mapstructure:"rpc-port"`
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

const (
	RepositoryNameOpGeth    RepositoryName = "op-geth"
	RepositoryNameOptimism  RepositoryName = "optimism"
	RepositoryNamePublisher RepositoryName = "publisher"

	L2ChainNameRollupA L2ChainName = "rollup-a"
	L2ChainNameRollupB L2ChainName = "rollup-b"
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

	requiredRepos := []RepositoryName{RepositoryNameOpGeth, RepositoryNameOptimism, RepositoryNamePublisher}
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

	rollupA, hasRollupA := c.ChainConfigs[L2ChainNameRollupA]
	rollupB, hasRollupB := c.ChainConfigs[L2ChainNameRollupB]

	if !hasRollupA {
		errs = append(errs, errors.New("l2.chain-configs.rollup-a is required"))
	} else {
		if rollupA.ID == 0 {
			errs = append(errs, errors.New("l2.chain-configs.rollup-a.id is required"))
		}
		if rollupA.RPCPort == 0 {
			errs = append(errs, errors.New("l2.chain-configs.rollup-a.rpc-port is required"))
		}
	}

	if !hasRollupB {
		errs = append(errs, errors.New("l2.chain-configs.rollup-b is required"))
	} else {
		if rollupB.ID == 0 {
			errs = append(errs, errors.New("l2.chain-configs.rollup-b.id is required"))
		}
		if rollupB.RPCPort == 0 {
			errs = append(errs, errors.New("l2.chain-configs.rollup-b.rpc-port is required"))
		}
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
