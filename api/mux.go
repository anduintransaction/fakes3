package api

import (
	"net/http"

	"github.com/anduintransaction/fakes3/config"
	"github.com/anduintransaction/fakes3/datastore"
	goji "goji.io"
	"goji.io/pat"
)

// Server for s3 Api
type Server struct {
	Mux           http.Handler
	config        *config.Config
	partStorage   *datastore.PartStorage
	objectStorage *datastore.ObjectStorage
}

// NewServer returns a new S3 Api Server
func NewServer(config *config.Config) *Server {
	s := &Server{}
	s.config = config
	s.Mux = s.newMux()
	s.partStorage = datastore.NewPartStorage(s.config.S3ApiServer.DataFolder)
	s.objectStorage = datastore.NewObjectStorage(s.config.S3ApiServer.DataFolder)
	return s
}

// Mux returns the HTTP handler for s3 Api
func (s *Server) newMux() http.Handler {
	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/:bucket/*"), s.getObjectRoute)
	mux.HandleFunc(pat.Post("/:bucket/*"), s.postObjectRoute)
	mux.HandleFunc(pat.Put("/:bucket/*"), s.putObjectRoute)
	mux.HandleFunc(pat.Delete("/:bucket/*"), s.deleteObjectRoute)
	return mux
}
