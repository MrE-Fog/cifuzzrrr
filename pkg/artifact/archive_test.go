package artifact_test

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code-intelligence.com/cifuzz/pkg/artifact"
	"code-intelligence.com/cifuzz/util/fileutil"
)

func TestWriteArchive(t *testing.T) {
	testdataDir := filepath.Join("testdata", "archive_test")
	require.DirExists(t, testdataDir)
	dir, err := os.MkdirTemp("", "write-archive-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	err = copy.Copy(testdataDir, dir)
	require.NoError(t, err)

	// Create an empty directory to test that WriteArchive handles it - it can't be kept in testdata since Git doesn't
	// allow checking in empty directories.
	err = os.MkdirAll(filepath.Join(dir, "empty_dir"), 0755)
	require.NoError(t, err)

	// Walk the testdata dir and add all contents to a fileMap for WriteArchive.
	fileMap := make(artifact.FileMap)
	err = artifact.AddDirToFileMap(fileMap, "", dir)
	require.NoError(t, err)

	archive, err := os.CreateTemp("", "artifact-*.tar.gz")
	require.NoError(t, err)
	defer archive.Close()
	err = artifact.WriteArchive(archive, fileMap)
	require.NoError(t, err)
	err = archive.Sync()
	require.NoError(t, err)

	// Unpack archive contents with tar.
	out, err := os.MkdirTemp("", "archive-test-*")
	require.NoError(t, err)
	cmd := exec.Command("tar", "-xvf", archive.Name(), "-C", out)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	require.NoError(t, err)

	remainingExpectedEntries := []struct {
		RelPath          string
		FileContent      string
		IsExecutableFile bool
	}{
		{".", "", false},
		{"dir1", "", false},
		{filepath.Join("dir1", "symlink"), "#!/usr/bin/env bash", true},
		{filepath.Join("dir1", "dir2"), "", false},
		{filepath.Join("dir1", "dir2", "test.sh"), "#!/usr/bin/env bash", true},
		{filepath.Join("dir1", "dir2", "test.txt"), "foobar", false},
		{"empty_dir", "", false},
	}
	// Verify that the archive contains exactly the expected files and directories.
	// Do not assert group and other permissions which may be affected by masks.
	err = filepath.WalkDir(out, func(absPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(out, absPath)
		if err != nil {
			return err
		}
		for i, expectedEntry := range remainingExpectedEntries {
			if relPath != expectedEntry.RelPath {
				continue
			}

			shouldBeDir := expectedEntry.FileContent == ""
			isDir := fileutil.IsDir(absPath)
			assert.Equalf(t, shouldBeDir, isDir, "Directory/file status doesn't match for %q", relPath)

			if isDir {
				remainingExpectedEntries = append(remainingExpectedEntries[:i], remainingExpectedEntries[i+1:]...)
				return nil
			}

			// Perform additional checks on files.
			stat, err := os.Lstat(absPath)
			require.NoError(t, err)
			assert.Falsef(
				t,
				stat.Mode()&os.ModeSymlink == os.ModeSymlink,
				"Expected symlinks to be archived as regular files: %q is a symlink",
				relPath,
			)

			if runtime.GOOS != "windows" {
				shouldBeExecutable := expectedEntry.IsExecutableFile
				isExecutable := stat.Mode()&0100 == 0100
				assert.Equalf(
					t,
					shouldBeExecutable,
					isExecutable,
					"Expected executable bit to be preserved, unexpected value for %s",
					relPath,
				)
			}

			content, err := os.ReadFile(absPath)
			require.NoError(t, err)
			assert.Equalf(t, expectedEntry.FileContent, string(content), "Contents are not as expected: %q", relPath)

			remainingExpectedEntries = append(remainingExpectedEntries[:i], remainingExpectedEntries[i+1:]...)
			return nil
		}
		assert.Fail(t, "Unexpected archive content: "+relPath)
		return nil
	})
	require.NoError(t, err)
	var msg strings.Builder
	for _, missingEntry := range remainingExpectedEntries {
		msg.WriteString(fmt.Sprintf("  %q\n", missingEntry.RelPath))
	}
	require.Empty(t, remainingExpectedEntries, "Archive did not contain the following expected entries: %s", msg.String())
}
