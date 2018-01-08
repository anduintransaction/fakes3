package api

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"goji.io/pat"
	"goji.io/pattern"
)

func (s *Server) getObjectRoute(w http.ResponseWriter, r *http.Request) {
	bucket := pat.Param(r, "bucket")
	objectKey := extractObjectKeyFromPath(pattern.Path(r.Context()))
	logrus.Debugf("Getting object %q from bucket %q", objectKey, bucket)
	objectPath := s.objectStorage.GetObjectFilePath(bucket, objectKey)
	if objectPath == "" {
		logrus.Warnf("Object not found or is not a file: %q", objectKey)
		notFoundResponse(w, r)
		return
	}
	addCORSHeaders(w)
	w.Header().Add("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, objectPath)
}

func (s *Server) postObjectRoute(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	switch {
	case queryKeyExists(queryParams, "uploads"):
		s.initializeMultipartUpload(w, r)
	case queryKeyExists(queryParams, "uploadId"):
		s.completeMultipartUpload(w, r)
	default:
		notFoundResponse(w, r)
	}
}

func (s *Server) putObjectRoute(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	switch {
	case queryKeyExists(queryParams, "partNumber") && queryKeyExists(queryParams, "uploadId"):
		s.uploadPart(w, r)
	case len(queryParams) == 0:
		s.normalUpload(w, r)
	default:
		notFoundResponse(w, r)
	}
}

func (s *Server) deleteObjectRoute(w http.ResponseWriter, r *http.Request) {
	bucket := pat.Param(r, "bucket")
	objectKey := extractObjectKeyFromPath(pattern.Path(r.Context()))
	logrus.Debugf("Deleting object %q from bucket %q", objectKey, bucket)
	err := s.objectStorage.DeleteObject(bucket, objectKey)
	if err != nil {
		logrus.Error(err)
		errorResponse(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	writeEmptySuccessResponse(w)
}

func (s *Server) optionObjectRoute(w http.ResponseWriter, r *http.Request) {
	addCORSHeaders(w)
	writeEmptySuccessResponse(w)
}
