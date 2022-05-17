package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func GetBlobs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	digest := vars["digest"]
	logrus.Info("GetBlobs: ", name+":"+digest)
}
