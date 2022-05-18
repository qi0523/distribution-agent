package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func ServeBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	digest := vars["digest"]
	logrus.Info("GetBlobs: ", name+":"+digest)
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
}
