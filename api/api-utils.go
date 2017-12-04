package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/palantir/stacktrace"
	uuid "github.com/satori/go.uuid"
)

const (
	defaultResponseNamespace = "http://s3.amazonaws.com/doc/2006-03-01/"
	defaultXMLContentType    = "application/xml"
)

type xmlErrorResponse struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	RequestID string   `xml:"RequestId"`
	HostID    string   `xml:"HostId"`
}

func queryKeyExists(params url.Values, key string) bool {
	_, ok := params[key]
	return ok
}

func dumpRequest(r *http.Request) string {
	dump, _ := httputil.DumpRequest(r, true)
	return string(dump)
}

func generateETag(content []byte) string {
	hasher := md5.New()
	hasher.Write(content)
	return hex.EncodeToString(hasher.Sum(nil))
}

func generateRandomETag() string {
	return "\"" + generateETag(uuid.NewV4().Bytes()) + "\""
}

func notFoundResponse(w http.ResponseWriter, r *http.Request) {
	writeCommonHeaders(w.Header())
	http.NotFound(w, r)
}

func errorResponse(w http.ResponseWriter) {
	writeXMLErrorResponse(w, http.StatusInternalServerError, "InternalError", "We encountered an internal error. Please try again.")
}

func writeXMLErrorResponse(w http.ResponseWriter, statusCode int, code, message string) error {
	requestID := uuid.NewV4().String()
	responseHeader := w.Header()
	responseHeader.Add("Server", "AmazonS3")
	responseHeader.Add("Date", time.Now().Format(http.TimeFormat))
	responseHeader.Add("x-amz-id-2", uuid.NewV4().String())
	responseHeader.Add("x-amz-request-id", requestID)

	errXML := &xmlErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
		HostID:    "fakes3",
	}
	content, err := xml.MarshalIndent(errXML, "", "    ")
	if err != nil {
		return stacktrace.Propagate(err, "Cannot marshal xml")
	}
	w.WriteHeader(statusCode)
	_, err = w.Write(content)
	return stacktrace.Propagate(err, "Cannot write response")
}

func writeCommonHeaders(responseHeader http.Header) {
	responseHeader.Add("Server", "AmazonS3")
	responseHeader.Add("Date", time.Now().Format(http.TimeFormat))
	responseHeader.Add("x-amz-id-2", uuid.NewV4().String())
	responseHeader.Add("x-amz-request-id", uuid.NewV4().String())
}

func writeEmptySuccessResponse(w http.ResponseWriter) error {
	writeCommonHeaders(w.Header())
	_, err := w.Write([]byte{})
	return stacktrace.Propagate(err, "Cannot write response")
}

func writeXMLResponse(w http.ResponseWriter, response interface{}) error {
	writeCommonHeaders(w.Header())
	content, err := xml.Marshal(response)
	if err != nil {
		return stacktrace.Propagate(err, "Cannot marshal response to xml")
	}
	w.Header().Add("Content-Type", defaultXMLContentType)
	_, err = w.Write(content)
	return stacktrace.Propagate(err, "Cannot write response")
}

func addCORSHeaders(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Methods", "*")
	w.Header().Add("Access-Control-Expose-Headers", "ETag, Accept-Ranges, Content-Range, Content-Encoding, Content-Length")
	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Allow-Origin", "*")
}

func extractObjectKeyFromPath(path string) string {
	unescapedPath, _ := url.PathUnescape(path)
	return strings.TrimPrefix(unescapedPath, "/")
}

func calculateAdvertiseAddress(advertisedAddress string, r *http.Request) string {
	if advertisedAddress != "" {
		return advertisedAddress
	}
	return fmt.Sprintf("http://%s", r.Host)
}

func generateFullObjectPath(advertisedAddress string, r *http.Request, bucket, objectKey string) string {
	advertisedAddress = calculateAdvertiseAddress(advertisedAddress, r)
	return fmt.Sprintf("%s/%s/%s", advertisedAddress, escapePath(bucket), escapePath(objectKey))
}

func escapePath(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
