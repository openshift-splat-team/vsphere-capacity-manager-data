package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

// IsManifestDirEmpty determines if the provided directory is empty
func IsManifestDirEmpty(manifestDir string) (bool, error) {
	entries, err := os.ReadDir(manifestDir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

// WriteManifest writes a manifest to the manifestDir
func WriteManifest(v any, manifestDir, fileName string) error {
	path := filepath.Join(manifestDir, fileName)

	marshalled, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("error while marshalling manifest %s: %w", fileName, err)
	}

	return os.WriteFile(path, marshalled, 0644)
}
