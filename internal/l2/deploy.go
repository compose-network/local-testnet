package l2

import (
    "encoding/json"
    "fmt"
    "log/slog"
    "os"
    "path/filepath"
    "strings"

    "github.com/compose-network/local-testnet/configs"
    "github.com/compose-network/local-testnet/internal/l2/infra/docker"
    l2path "github.com/compose-network/local-testnet/internal/l2/path"
    "github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
    Use:   "deploy [op-geth|publisher|all]",
    Short: "Build and restart selected L2 services for rapid local development",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        target := strings.ToLower(args[0])
        allowed := map[string]bool{"op-geth": true, "publisher": true, "all": true}
        if !allowed[target] {
            return fmt.Errorf("invalid service '%s' (expected: op-geth|publisher|all)", target)
        }

        rootDir, err := os.Getwd()
        if err != nil {
            return fmt.Errorf("failed to get working directory: %w", err)
        }
        localnetDir := filepath.Join(rootDir, localnetDirName)
        networksDir := filepath.Join(localnetDir, networksDirName)
        servicesDir := filepath.Join(localnetDir, servicesDirName)

        composePath, err := docker.EnsureComposeFile(localnetDir)
        if err != nil {
            return fmt.Errorf("failed to prepare docker-compose file: %w", err)
        }

        env, err := buildDevComposeEnv(rootDir, networksDir, servicesDir, configs.Values.L2)
        if err != nil {
            return err
        }

        services := mapServices(target)
        slog.With("services", services).Info("building services from local sources")
        if err := docker.ComposeBuild(cmd.Context(), composePath, env, services...); err != nil {
            return fmt.Errorf("failed to build services: %w", err)
        }

        slog.Info("restarting services to apply new images")
        if err := docker.ComposeRestart(cmd.Context(), composePath, env, services...); err != nil {
            return fmt.Errorf("failed to restart services: %w", err)
        }

        slog.Info("deploy completed successfully")
        return nil
    },
}

func mapServices(target string) []string {
    switch target {
    case "op-geth":
        return []string{"op-geth-a", "op-geth-b"}
    case "publisher":
        return []string{"publisher"}
    default:
        return []string{"publisher", "op-geth-a", "op-geth-b"}
    }
}

func buildDevComposeEnv(rootDir, networksDir, servicesDir string, cfg configs.L2) (map[string]string, error) {
    env := make(map[string]string)

    // Resolve build contexts (prefer local-path; otherwise cloned path)
    publisherPath := cfg.Repositories[configs.RepositoryNamePublisher].LocalPath
    if publisherPath == "" {
        publisherPath = filepath.Join(servicesDir, string(configs.RepositoryNamePublisher))
    } else {
        publisherPath = expandUserHome(publisherPath)
    }
    opGethPath := cfg.Repositories[configs.RepositoryNameOpGeth].LocalPath
    if opGethPath == "" {
        opGethPath = filepath.Join(servicesDir, string(configs.RepositoryNameOpGeth))
    } else {
        opGethPath = expandUserHome(opGethPath)
    }

    rollupAConfigPath := filepath.Join(networksDir, string(configs.L2ChainNameRollupA))
    rollupBConfigPath := filepath.Join(networksDir, string(configs.L2ChainNameRollupB))

    // Convert volume mount paths to host paths
    rollupAHost, err := l2path.GetHostPath(rollupAConfigPath)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve host path for rollup-a config: %w", err)
    }
    rollupBHost, err := l2path.GetHostPath(rollupBConfigPath)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve host path for rollup-b config: %w", err)
    }
    rootHost, err := l2path.GetHostPath(rootDir)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve host path for rootDir: %w", err)
    }

    env["ROOT_DIR"] = rootHost
    env["WALLET_PRIVATE_KEY"] = cfg.Wallet.PrivateKey
    env["WALLET_ADDRESS"] = cfg.Wallet.Address
    env["L1_EL_URL"] = cfg.L1ElURL
    env["L1_CL_URL"] = cfg.L1ClURL
    env["L1_CHAIN_ID"] = fmt.Sprintf("%d", cfg.L1ChainID)
    // Ensure compose network name is always passed to publisher
    env["COMPOSE_NETWORK_NAME"] = cfg.ComposeNetworkName
    env["COORDINATOR_PRIVATE_KEY"] = cfg.CoordinatorPrivateKey
    env["SEQUENCER_PRIVATE_KEY"] = cfg.CoordinatorPrivateKey
    env["SP_L1_SUPERBLOCK_CONTRACT"] = ""

    env["PUBLISHER_PATH"] = publisherPath
    env["OP_GETH_PATH"] = opGethPath

    env["ROLLUP_A_CHAIN_ID"] = fmt.Sprintf("%d", cfg.ChainConfigs[configs.L2ChainNameRollupA].ID)
    env["ROLLUP_A_RPC_PORT"] = fmt.Sprintf("%d", cfg.ChainConfigs[configs.L2ChainNameRollupA].RPCPort)
    env["ROLLUP_A_CONFIG_PATH"] = rollupAHost
    env["ROLLUP_A_CONFIG_PATH_CONTAINER"] = filepath.Join(networksDir, string(configs.L2ChainNameRollupA))

    env["ROLLUP_B_CHAIN_ID"] = fmt.Sprintf("%d", cfg.ChainConfigs[configs.L2ChainNameRollupB].ID)
    env["ROLLUP_B_RPC_PORT"] = fmt.Sprintf("%d", cfg.ChainConfigs[configs.L2ChainNameRollupB].RPCPort)
    env["ROLLUP_B_CONFIG_PATH"] = rollupBHost
    env["ROLLUP_B_CONFIG_PATH_CONTAINER"] = filepath.Join(networksDir, string(configs.L2ChainNameRollupB))

    env["SP_L1_DISPUTE_GAME_FACTORY"] = "" // unchanged for dev deploy

    env["OP_BATCHER_IMAGE_TAG"] = cfg.Images[configs.ImageNameOpBatcher].Tag
    env["OP_NODE_IMAGE_TAG"] = cfg.Images[configs.ImageNameOpNode].Tag
    env["OP_PROPOSER_IMAGE_TAG"] = cfg.Images[configs.ImageNameOpProposer].Tag

    // Best-effort: preserve mailbox addresses for op-geth restarts by reading
    // the addresses written during the initial deployment. If files are absent,
    // leave env unset (compose service tolerates empty defaults).
    type contractFile struct {
        Addresses map[string]string `json:"addresses"`
    }
    readMailbox := func(path string) string {
        data, err := os.ReadFile(path)
        if err != nil {
            return ""
        }
        var cf contractFile
        if err := json.Unmarshal(data, &cf); err != nil {
            return ""
        }
        if v := cf.Addresses["Mailbox"]; strings.TrimSpace(v) != "" {
            return v
        }
        return ""
    }

    ma := readMailbox(filepath.Join(networksDir, string(configs.L2ChainNameRollupA), "contracts.json"))
    mb := readMailbox(filepath.Join(networksDir, string(configs.L2ChainNameRollupB), "contracts.json"))
    if ma != "" {
        env["MAILBOX_A"] = ma
    }
    if mb != "" {
        env["MAILBOX_B"] = mb
    }

    return env, nil
}

func expandUserHome(p string) string {
    if p == "" || p[0] != '~' {
        return p
    }
    home, err := os.UserHomeDir()
    if err != nil {
        return p
    }
    if p == "~" {
        return home
    }
    return filepath.Join(home, p[2:])
}
