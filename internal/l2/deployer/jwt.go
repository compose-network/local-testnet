package deployer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// GenerateJWTSecrets generates JWT secret files for both rollups.
func GenerateJWTSecrets(networksDir string) error {
	slog.Info("generating JWT secrets")

	rollupADir := filepath.Join(networksDir, "rollup-a")
	rollupBDir := filepath.Join(networksDir, "rollup-b")

	if err := generateJWTSecretForChain(rollupADir); err != nil {
		return fmt.Errorf("failed to generate JWT secret for rollup-a: %w", err)
	}

	if err := generateJWTSecretForChain(rollupBDir); err != nil {
		return fmt.Errorf("failed to generate JWT secret for rollup-b: %w", err)
	}

	slog.Info("JWT secrets generated successfully")

	return nil
}

func generateJWTSecretForChain(networkDir string) error {
	jwtPath := filepath.Join(networkDir, "jwt.txt")

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as hex string with 0x prefix
	jwtSecret := "0x" + hex.EncodeToString(secret)

	if err := os.WriteFile(jwtPath, []byte(jwtSecret), 0600); err != nil {
		return fmt.Errorf("failed to write JWT secret file: %w", err)
	}

	slog.Info("JWT secret file written", "path", jwtPath)

	return nil
}
