package api

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"goji.io/pat"
	"goji.io/pattern"
)

func (s *Server) normalUpload(w http.ResponseWriter, r *http.Request) {
	bucket := pat.Param(r, "bucket")
	objectKey := extractObjectKeyFromPath(pattern.Path(r.Context()))
	logrus.Debugf("Uploading object %q to bucket %q", objectKey, bucket)
	err := s.objectStorage.PutObject(bucket, objectKey, r.Body)
	if err != nil {
		logrus.Error(err)
		errorResponse(w)
		return
	}
	writeEmptySuccessResponse(w)
}
