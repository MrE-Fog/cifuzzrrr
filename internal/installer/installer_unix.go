//go:build !windows

package installer

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"code-intelligence.com/cifuzz/pkg/log"
)

func RegisterCMakePackage(packageDir string) error {
	// Install the CMake package for the current user only by registering it
	// with the user package registry. This requires creating a file
	// ~/.cmake/packages/CIFuzz/CIFuzz containing the path to the root directory
	// of the CMake integration.
	// See https://cmake.org/cmake/help/latest/manual/cmake-packages.7.html#user-package-registry
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.WithStack(err)
	}
	cmakePackagesDir := filepath.Join(homeDir, ".cmake", "packages", "cifuzz")
	log.Printf("Adding CMake package to %s", cmakePackagesDir)
	err = os.MkdirAll(cmakePackagesDir, 0755)
	if err != nil {
		return errors.WithStack(err)
	}
	err = os.WriteFile(filepath.Join(cmakePackagesDir, "cifuzz"), []byte(packageDir), 0644)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
