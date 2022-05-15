package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
)

func GetManifest(w http.ResponseWriter, r *http.Request) {
	logrus.Info("ctx: ", r.Context())
	vars := mux.Vars(r)
	name := vars["name"]
	ref := vars["reference"]
	dgst, err := digest.Parse(ref)
	if err != nil { //tag

	}
}
