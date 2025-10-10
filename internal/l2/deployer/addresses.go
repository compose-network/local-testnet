package deployer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// ExportAddresses exports L1 contract addresses for both rollups.
func ExportAddresses(stateDir, networksDir string, rollupAChainID, rollupBChainID int) error {
	slog.Info("exporting L1 contract addresses")

	statePath := filepath.Join(stateDir, "state.json")
	stateData, err := os.ReadFile(statePath)
	if err != nil {
		return fmt.Errorf("failed to read state.json: %w", err)
	}

	var state map[string]interface{}
	if err := json.Unmarshal(stateData, &state); err != nil {
		return fmt.Errorf("failed to parse state.json: %w", err)
	}

	// Build deployment map by chain ID
	deployments := make(map[string]map[string]interface{})
	if opChainDeployments, ok := state["opChainDeployments"].([]interface{}); ok {
		for _, entry := range opChainDeployments {
			if deployment, ok := entry.(map[string]interface{}); ok {
				if id, ok := deployment["id"].(string); ok {
					deployments[strings.ToLower(id)] = deployment
				}
			}
		}
	}

	rollupADir := filepath.Join(networksDir, "rollup-a")
	rollupBDir := filepath.Join(networksDir, "rollup-b")

	// Export addresses for Rollup A
	if err := exportAddressesForChain(rollupADir, rollupAChainID, deployments); err != nil {
		return fmt.Errorf("failed to export addresses for rollup-a: %w", err)
	}

	// Export addresses for Rollup B
	if err := exportAddressesForChain(rollupBDir, rollupBChainID, deployments); err != nil {
		return fmt.Errorf("failed to export addresses for rollup-b: %w", err)
	}

	slog.Info("L1 contract addresses exported successfully")
	return nil
}

func exportAddressesForChain(networkDir string, chainID int, deployments map[string]map[string]interface{}) error {
	chainIDHex := strings.ToLower(chainIDToHex(chainID))
	deployment, ok := deployments[chainIDHex]
	if !ok {
		return fmt.Errorf("no deployment found for chain ID %d", chainID)
	}

	// Map state.json keys to output labels
	labelMap := map[string]string{
		"optimismPortalProxyAddress":     "OPTIMISM_PORTAL",
		"l1StandardBridgeProxyAddress":   "L1_STANDARD_BRIDGE",
		"systemConfigProxyAddress":       "SYSTEM_CONFIG",
		"L2OutputOracleProxyAddress":     "L2_OUTPUT_ORACLE",
		"disputeGameFactoryProxyAddress": "DISPUTE_GAME_FACTORY",
	}

	addresses := make(map[string]string)
	for key, label := range labelMap {
		if value, ok := deployment[key].(string); ok && value != "" {
			addresses[label] = value
		}
	}

	// Write addresses.json
	addressesPath := filepath.Join(networkDir, "addresses.json")
	addressesData, err := json.MarshalIndent(addresses, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal addresses.json: %w", err)
	}

	if err := os.MkdirAll(networkDir, 0755); err != nil {
		return fmt.Errorf("failed to create network directory: %w", err)
	}

	if err := os.WriteFile(addressesPath, append(addressesData, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write addresses.json: %w", err)
	}

	slog.Info("addresses.json written", "path", addressesPath)

	// Write runtime.env for op-proposer
	if err := writeRuntimeEnv(networkDir, deployment); err != nil {
		return fmt.Errorf("failed to write runtime.env: %w", err)
	}

	return nil
}

func writeRuntimeEnv(networkDir string, deployment map[string]interface{}) error {
	runtimeEnvPath := filepath.Join(networkDir, "runtime.env")

	var lines []string

	// Add L2 Output Oracle addresses if available
	if l2oo := getAddress(deployment, "L2OutputOracleProxyAddress"); l2oo != "" {
		lines = append(lines, fmt.Sprintf("L2OO_ADDRESS=%s", l2oo))
		lines = append(lines, fmt.Sprintf("OP_PROPOSER_L2OO_ADDRESS=%s", l2oo))
	}

	// Add Dispute Game Factory addresses if available
	if dgf := getAddress(deployment, "disputeGameFactoryProxyAddress"); dgf != "" {
		lines = append(lines, fmt.Sprintf("DISPUTE_GAME_FACTORY_ADDRESS=%s", dgf))
		lines = append(lines, fmt.Sprintf("OP_PROPOSER_GAME_FACTORY_ADDRESS=%s", dgf))
	}

	content := strings.Join(lines, "\n") + "\n"

	if err := os.WriteFile(runtimeEnvPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write runtime.env: %w", err)
	}

	slog.Info("runtime.env written", "path", runtimeEnvPath)

	return nil
}

func getAddress(deployment map[string]interface{}, key string) string {
	if value, ok := deployment[key].(string); ok && value != "" {
		return value
	}

	return ""
}
