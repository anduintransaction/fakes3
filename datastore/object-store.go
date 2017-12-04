package datastore

import (
	"io"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/palantir/stacktrace"
)

// ObjectStorage stores s3 objects
type ObjectStorage struct {
	objectStorageFolder string
	tmpFolder           string
}

// NewObjectStorage returns new ObjectStorage
func NewObjectStorage(s3DataFolder string) *ObjectStorage {
	return &ObjectStorage{
		objectStorageFolder: filepath.Join(s3DataFolder, "objects"),
		tmpFolder:           filepath.Join(s3DataFolder, "tmp"),
	}
}

// MergeParts merges upload parts to create a new object
func (o *ObjectStorage) MergeParts(bucket, objectKey, uploadID string, partStorage *PartStorage) error {
	objectTmpPath := filepath.Join(o.tmpFolder, bucket, objectKey)
	err := createParentDirForFile(objectTmpPath)
	if err != nil {
		return err
	}
	w, err := os.Create(objectTmpPath)
	if err != nil {
		return stacktrace.Propagate(err, "Cannot create object tmp path %q", objectTmpPath)
	}
	err = partStorage.MergeParts(uploadID, w)
	if err != nil {
		w.Close()
		os.Remove(objectTmpPath)
		return err
	}
	objectPath := filepath.Join(o.objectStorageFolder, bucket, objectKey)
	err = createParentDirForFile(objectPath)
	if err != nil {
		os.Remove(objectTmpPath)
		return err
	}
	err = os.Rename(objectTmpPath, objectPath)
	if err != nil {
		os.Remove(objectTmpPath)
		return stacktrace.Propagate(err, "Cannot move object tmp path %q", objectTmpPath)
	}
	logrus.Debugf("Successfully merged object to %q", objectPath)
	return nil
}

// PutObject stores an object
func (o *ObjectStorage) PutObject(bucket, objectKey string, source io.Reader) error {
	objectPath := filepath.Join(o.objectStorageFolder, bucket, objectKey)
	err := createParentDirForFile(objectPath)
	if err != nil {
		return err
	}
	w, err := os.Create(objectPath)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, source)
	return stacktrace.Propagate(err, "Cannot store object %q to bucket %q", objectKey, bucket)
}

// DeleteObject deletes an object
func (o *ObjectStorage) DeleteObject(bucket, objectKey string) error {
	objectPath := filepath.Join(o.objectStorageFolder, bucket, objectKey)
	_, err := os.Stat(objectPath)
	if err != nil {
		return nil
	}
	err = os.Remove(objectPath)
	return stacktrace.Propagate(err, "Cannot delete object %q from bucket %q", objectKey, bucket)
}

// GetObjectFilePath returns absolute path to a object. Will return empty string if file not found or is not a file
func (o *ObjectStorage) GetObjectFilePath(bucket, objectKey string) string {
	objectPath := filepath.Join(o.objectStorageFolder, bucket, objectKey)
	info, err := os.Stat(objectPath)
	if err != nil {
		return ""
	}
	if info.IsDir() {
		return ""
	}
	return objectPath
}
