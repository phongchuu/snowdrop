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
	dirPath := filepath.Join(getWorkspaceDir(), "console", "resources", "trans")
	files, _ := os.ReadDir(dirPath)

	virtualFileSys := fstest.MapFS{}

	for _, file := range files {
		filePath := filepath.Join(dirPath, file.Name())
		virtualPath := filepath.Join("resources", "trans", file.Name())
		data, _ := os.ReadFile(filePath)

		virtualFileSys[virtualPath] = &fstest.MapFile{Data: data}
	}

	return virtualFileSys
}
