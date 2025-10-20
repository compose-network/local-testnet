package dispute

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/logger"
)

// Service handles dispute game factory deployment
type Service struct {
	rootDir      string
	contractsDir string // Path to cloned compose-contracts repo
	deployerPK   string
	cfg          configs.L2
	logger       *slog.Logger
}

// NewService creates a new dispute deployment service
func NewService(rootDir string, cfg configs.L2) *Service {
	// Contracts are located in the L1-settlement subdirectory
	contractsDir := filepath.Join(rootDir, "internal", "l2", "services", string(configs.RepositoryNameComposeContracts), "L1-settlement")

	return &Service{
		rootDir:      rootDir,
		contractsDir: contractsDir,
		deployerPK:   cfg.Wallet.PrivateKey,
		cfg:          cfg,
		logger:       logger.Named("dispute_deployer"),
	}
}

// Deploy executes the full deployment workflow and returns DisputeGameFactory proxy address
func (s *Service) Deploy(ctx context.Context) (string, error) {
	s.logger.Info("starting dispute contracts deployment")

	if _, err := os.Stat(s.contractsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("L1-settlement directory not found at %s. Make sure compose-contracts repository is cloned first", s.contractsDir)
	}

	justfilePath := filepath.Join(s.contractsDir, "justfile")
	if _, err := os.Stat(justfilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("justfile not found in %s. The compose-contracts/L1-settlement may not be properly set up", s.contractsDir)
	}

	s.logger.Info("generating networks.toml")
	if err := s.generateNetworksToml(); err != nil {
		return "", fmt.Errorf("failed to generate networks.toml: %w", err)
	}

	s.logger.Info("generating .env file")
	if err := s.generateEnvFile(); err != nil {
		return "", fmt.Errorf("failed to generate .env file: %w", err)
	}

	s.logger.Info("running just setup")
	if err := s.runJustCommand(ctx, "setup"); err != nil {
		return "", fmt.Errorf("failed to run just setup: %w", err)
	}

	s.logger.Info("running just build")
	if err := s.runJustCommand(ctx, "build"); err != nil {
		return "", fmt.Errorf("failed to run just build: %w", err)
	}

	s.logger.Info("running just deploy-network localnet")
	if err := s.runJustCommand(ctx, "deploy-network", "localnet"); err != nil {
		return "", fmt.Errorf("failed to deploy to localnet: %w", err)
	}

	s.logger.Info("parsing deployments.json")
	addr, err := s.parseDisputeGameFactoryAddress()
	if err != nil {
		return "", fmt.Errorf("failed to parse DisputeGameFactory address: %w", err)
	}

	s.logger.With("address", addr).Info("dispute contracts deployed successfully")

	return addr, nil
}

// generateNetworksToml creates networks.toml from template and config
func (s *Service) generateNetworksToml() error {
	templatePath := filepath.Join(s.rootDir, "internal", "l2", "l1deployment", "dispute", "networks.tmpl")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template file: %w", err)
	}

	type templateData struct {
		NetworkName              string
		RpcURL                   string
		ChainID                  int
		ExplorerURL              string
		ExplorerAPIURL           string
		VerifierAddress          string
		OwnerAddress             string
		ProposerAddress          string
		AggregationVkey          string
		StartingSuperblockNumber int
		AdminAddress             string
	}

	data := templateData{
		NetworkName:              s.cfg.Dispute.NetworkName,
		RpcURL:                   s.cfg.L1ElURL,
		ChainID:                  s.cfg.L1ChainID,
		ExplorerURL:              s.cfg.Dispute.ExplorerURL,
		ExplorerAPIURL:           s.cfg.Dispute.ExplorerAPIURL,
		VerifierAddress:          s.cfg.Dispute.VerifierAddress,
		OwnerAddress:             s.cfg.Dispute.OwnerAddress,
		ProposerAddress:          s.cfg.Dispute.ProposerAddress,
		AggregationVkey:          s.cfg.Dispute.AggregationVkey,
		StartingSuperblockNumber: s.cfg.Dispute.StartingSuperblockNumber,
		AdminAddress:             s.cfg.Dispute.AdminAddress,
	}

	outputPath := filepath.Join(s.contractsDir, "networks.toml")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create networks.toml: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// generateEnvFile creates .env file from template
func (s *Service) generateEnvFile() error {
	templatePath := filepath.Join(s.rootDir, "internal", "l2", "l1deployment", "dispute", "env.tmpl")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse env template file: %w", err)
	}

	data := struct {
		DeployerPrivateKey string
	}{
		DeployerPrivateKey: s.deployerPK,
	}

	envPath := filepath.Join(s.contractsDir, ".env")
	file, err := os.Create(envPath)
	if err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute env template: %w", err)
	}

	if err := os.Chmod(envPath, 0600); err != nil {
		return fmt.Errorf("failed to set .env file permissions: %w", err)
	}

	return nil
}

// runJustCommand executes a just command in the contracts directory
func (s *Service) runJustCommand(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "just", args...)
	cmd.Dir = s.contractsDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	s.logger.
		With("command", fmt.Sprintf("just %s", strings.Join(args, " "))).
		With("working_dir", s.contractsDir).
		Info("executing just command")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command 'just %s' failed in directory %s: %w", strings.Join(args, " "), s.contractsDir, err)
	}

	return nil
}

// parseDisputeGameFactoryAddress reads deployments.json and extracts DisputeGameFactory proxy address
func (s *Service) parseDisputeGameFactoryAddress() (string, error) {
	deploymentsPath := filepath.Join(s.contractsDir, "deployments.json")

	data, err := os.ReadFile(deploymentsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read deployments.json: %w", err)
	}

	var deployments map[string]struct {
		DisputeGameFactory struct {
			Proxy string `json:"proxy"`
		} `json:"DisputeGameFactory"`
	}

	if err := json.Unmarshal(data, &deployments); err != nil {
		return "", fmt.Errorf("failed to parse deployments.json: %w", err)
	}

	localnet, ok := deployments["localnet"]
	if !ok {
		return "", fmt.Errorf("localnet deployment not found in deployments.json")
	}

	if localnet.DisputeGameFactory.Proxy == "" {
		return "", fmt.Errorf("DisputeGameFactory proxy address is empty")
	}

	return localnet.DisputeGameFactory.Proxy, nil
}
