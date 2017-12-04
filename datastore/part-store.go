package datastore

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/palantir/stacktrace"
)

// PartStorage stores multipart upload parts
type PartStorage struct {
	partStorageFolder string
}

// NewPartStorage returns new PartStorage
func NewPartStorage(s3DataFolder string) *PartStorage {
	return &PartStorage{
		partStorageFolder: filepath.Join(s3DataFolder, "parts"),
	}
}

// StorePart stores a part to the storage
func (ps *PartStorage) StorePart(uploadID string, partNumber int, source io.Reader) error {
	uploadFolder := filepath.Join(ps.partStorageFolder, uploadID)
	err := os.MkdirAll(uploadFolder, 0755)
	if err != nil {
		return stacktrace.Propagate(err, "Cannot create upload folder %q", uploadFolder)
	}
	partFile := filepath.Join(uploadFolder, fmt.Sprintf("part-%d", partNumber))
	w, err := os.Create(partFile)
	if err != nil {
		return stacktrace.Propagate(err, "Cannot create part file %q", partFile)
	}
	defer w.Close()
	_, err = io.Copy(w, source)
	return stacktrace.Propagate(err, "Cannot write to part file %q", partFile)
}

// MergeParts merges all parts of an upload and write to a sink
func (ps *PartStorage) MergeParts(uploadID string, sink io.Writer) error {
	uploadFolder := filepath.Join(ps.partStorageFolder, uploadID)
	parts, err := ioutil.ReadDir(uploadFolder)
	if err != nil {
		return stacktrace.Propagate(err, "Cannot read upload folder %q", uploadFolder)
	}
	partNums := []int{}
	for _, part := range parts {
		partNum, err := strconv.ParseInt(strings.TrimPrefix(part.Name(), "part-"), 10, 64)
		if err != nil {
			return stacktrace.Propagate(err, "Invalid part name: %q", part.Name())
		}
		partNums = append(partNums, int(partNum))
	}
	sort.Ints(partNums)
	for _, partNum := range partNums {
		partFile := filepath.Join(uploadFolder, fmt.Sprintf("part-%d", partNum))
		r, err := os.Open(partFile)
		if err != nil {
			return stacktrace.Propagate(err, "Cannot open part file %q", partFile)
		}
		_, err = io.Copy(sink, r)
		if err != nil {
			return stacktrace.Propagate(err, "Cannot write to destination for part file %q", partFile)
		}
	}
	os.RemoveAll(uploadFolder)
	return nil
}
