package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	"github.com/qi0523/distribution-agent/storage"
	"github.com/sirupsen/logrus"
)

func ServeBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	ref := vars["digest"]
	logrus.Info("ServeBlob: ", name+":"+ref)
	dgst, err := digest.Parse(ref)
	if err != nil {
		logrus.Error("invalid digest: ", ref)
		return
	}
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
	if err := storage.GetBlobServer().ServeBlob(w, r, dgst); err != nil {
		logrus.Error("%v", err)
		return
	}
}
