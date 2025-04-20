package testutils

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing/fstest"
)

func getWorkspaceDir() string {
	cmd := exec.Command("go", "env", "GOWORK")
	output, _ := cmd.Output()

	return filepath.Dir(string(output))
}

// GetResourceFS creates a virtual file system (fs.FS).
func GetResourceFS() fs.FS {
	baseDir := filepath.Join(getWorkspaceDir(), "console", "resources")
	vfs := fstest.MapFS{}

	walkFn := func(path string, entry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case entry.IsDir():
			return nil
		default:
			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			virtualPath := filepath.Join("resources", relPath)
			vfs[virtualPath] = &fstest.MapFile{Data: content}

			return nil
		}
	}

	if err := filepath.WalkDir(baseDir, walkFn); err != nil {
		panic(err)
	}

	return vfs
}
