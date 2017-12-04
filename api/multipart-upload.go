package api

import (
	"encoding/xml"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	uuid "github.com/satori/go.uuid"
	"goji.io/pat"
	"goji.io/pattern"
)

type initializeMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Xmlns    string   `xml:"xmlns,attr"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	UploadID string   `xml:"UploadId"`
}

type completeMultipartUploadResult struct {
	XMLName  xml.Name `xml:"CompleteMultipartUploadResult"`
	Location string   `xml:"Location"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	ETag     string   `xml:"ETag"`
}

func (s *Server) initializeMultipartUpload(w http.ResponseWriter, r *http.Request) {
	bucket := pat.Param(r, "bucket")
	objectKey := extractObjectKeyFromPath(pattern.Path(r.Context()))
	logrus.Debugf("Initialize multipart upload to object %q, bucket %q", objectKey, bucket)
	err := writeXMLResponse(w, &initializeMultipartUploadResult{
		Xmlns:    defaultResponseNamespace,
		Bucket:   bucket,
		Key:      objectKey,
		UploadID: uuid.NewV4().String(),
	})
	if err != nil {
		logrus.Error(err)
		errorResponse(w)
	}
}

func (s *Server) uploadPart(w http.ResponseWriter, r *http.Request) {
	partNumber, err := strconv.ParseInt(r.URL.Query().Get("partNumber"), 10, 64)
	if err != nil {
		logrus.Error(err)
	}
	uploadID := r.URL.Query().Get("uploadId")
	logrus.Debugf("Got part %d from %q", partNumber, uploadID)
	err = s.partStorage.StorePart(uploadID, int(partNumber), r.Body)
	if err != nil {
		logrus.Error(err)
		errorResponse(w)
	}
	w.Header().Set("ETag", generateRandomETag())
	writeEmptySuccessResponse(w)
}

func (s *Server) completeMultipartUpload(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("uploadId")
	bucket := pat.Param(r, "bucket")
	objectKey := extractObjectKeyFromPath(pattern.Path(r.Context()))
	logrus.Debugf("Got complete multipart upload for %q, bucket %q, key %q", uploadID, bucket, objectKey)
	err := s.objectStorage.MergeParts(bucket, objectKey, uploadID, s.partStorage)
	if err != nil {
		logrus.Error(err)
		errorResponse(w)
	}
	etag := generateRandomETag()
	w.Header().Set("ETag", etag)
	err = writeXMLResponse(w, &completeMultipartUploadResult{
		Location: generateFullObjectPath(s.config.S3ApiServer.AdvertisedAddr, r, bucket, objectKey),
		Bucket:   bucket,
		Key:      objectKey,
		ETag:     etag,
	})
	if err != nil {
		logrus.Error(err)
		errorResponse(w)
	}
}
