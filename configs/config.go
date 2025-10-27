package configs

import (
	"errors"
	"fmt"
)

var Values Config

type (
	RepositoryName string
	L2ChainName    string
	ImageName      string

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
		ChainConfigs          map[L2ChainName]Chain         `mapstructure:"chain-configs"`
		Images                map[ImageName]Image           `mapstructure:"images"`
		DeploymentTarget      string                        `mapstructure:"deployment-target"`
		GenesisBalanceWei     string                        `mapstructure:"genesis-balance-wei"`
		OPDeployerVersion     string                        `mapstructure:"op-deployer-version"`
		Dispute               DisputeConfig                 `mapstructure:"dispute"`
	}

	DisputeConfig struct {
		NetworkName              string `mapstructure:"network-name"`
		ExplorerURL              string `mapstructure:"explorer-url"`
		ExplorerAPIURL           string `mapstructure:"explorer-api-url"`
		VerifierAddress          string `mapstructure:"verifier-address"`
		OwnerAddress             string `mapstructure:"owner-address"`
		ProposerAddress          string `mapstructure:"proposer-address"`
		AggregationVkey          string `mapstructure:"aggregation-vkey"`
		StartingSuperblockNumber int    `mapstructure:"starting-superblock-number"`
		AdminAddress             string `mapstructure:"admin-address"`
	}

	Chain struct {
		ID      int `mapstructure:"id"`
		RPCPort int `mapstructure:"rpc-port"`
	}

	Repository struct {
		URL    string `mapstructure:"url"`
		Branch string `mapstructure:"branch"`
	}

	Image struct {
		Tag string `mapstructure:"tag"`
	}

	Wallet struct {
		PrivateKey string `mapstructure:"private-key"`
		Address    string `mapstructure:"address"`
	}

	Observability struct {
	}
)

const (
	RepositoryNameOpGeth           RepositoryName = "op-geth"
	RepositoryNamePublisher        RepositoryName = "publisher"
	RepositoryNameComposeContracts RepositoryName = "compose-contracts"

	ImageNameOpNode     ImageName = "op-node"
	ImageNameOpProposer ImageName = "op-proposer"
	ImageNameOpBatcher  ImageName = "op-batcher"

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

	requiredRepos := []RepositoryName{RepositoryNameOpGeth, RepositoryNamePublisher}
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

	if c.OPDeployerVersion == "" {
		errs = append(errs, errors.New("l2.op-deployer-version is required"))
	}

	// Validate dispute config
	if c.Dispute.NetworkName == "" {
		errs = append(errs, errors.New("l2.dispute.network-name is required"))
	}
	if c.Dispute.VerifierAddress == "" {
		errs = append(errs, errors.New("l2.dispute.verifier-address is required"))
	}
	if c.Dispute.OwnerAddress == "" {
		errs = append(errs, errors.New("l2.dispute.owner-address is required"))
	}
	if c.Dispute.ProposerAddress == "" {
		errs = append(errs, errors.New("l2.dispute.proposer-address is required"))
	}
	if c.Dispute.AggregationVkey == "" {
		errs = append(errs, errors.New("l2.dispute.aggregation-vkey is required"))
	}
	if c.Dispute.AdminAddress == "" {
		errs = append(errs, errors.New("l2.dispute.admin-address is required"))
	}

	if len(errs) > 0 {
		return fmt.Errorf("L2 configuration validation failed: %w", errors.Join(errs...))
	}

	return nil
}
