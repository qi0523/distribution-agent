package v2

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	"github.com/qi0523/distribution-agent/handlers"
	"github.com/qi0523/distribution-agent/reference"
)

var (
	RouteNameManifest = "/v2/{name:" + reference.NameRegexp.String() + "}/manifests/{reference:" + reference.TagRegexp.String() + "|" + digest.DigestRegexp.String() + "}"
	RouteNameBlob     = "/v2/{name:" + reference.NameRegexp.String() + "}/blobs/{digest:" + digest.DigestRegexp.String() + "}"
)

func Router() *mux.Router {
	router := mux.NewRouter()

	router.StrictSlash(true)

	router.HandleFunc(RouteNameManifest, handlers.GetManifest).Methods(http.MethodHead)
	router.HandleFunc(RouteNameManifest, handlers.GetManifest).Methods(http.MethodGet)
	router.HandleFunc(RouteNameBlob, handlers.GetBlobs).Methods(http.MethodGet)

	return router
}
