package datastore

import (
	"os"
	"path/filepath"

	"github.com/palantir/stacktrace"
)

func createParentDirForFile(file string) error {
	parentDir := filepath.Dir(file)
	err := os.MkdirAll(parentDir, 0755)
	return stacktrace.Propagate(err, "Cannot create parent dir for %q", file)
}
