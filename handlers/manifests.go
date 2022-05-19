package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	"github.com/qi0523/distribution-agent/client"
	"github.com/sirupsen/logrus"
)

const (
	fileDir = "/var/lib/containerd/io.containerd.content.v1.content/blobs/sha256/"
)

func ResolvedManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	ref := vars["reference"]
	logrus.Info("ResolvedManifest: ", name+":"+ref)
	dgst, err := digest.Parse(ref)
	var mediaType = ""
	if err != nil { //tag
		mediaType, dgst, err = client.GetManifestInfoByTag(name, ref)
	} else {
		mediaType, err = client.GetManifestInfoByDigest(dgst)
	}
	if err != nil || mediaType == "" {
		return
	}

	fi, err := os.Stat(fileDir + dgst.String()[7:])
	if err != nil {
		return
	}
	logrus.Info("ResponseWriter@Content-length: ", fi.Size())
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Type", mediaType)
	w.Header().Set("Content-Length", fmt.Sprint(fi.Size()))
	w.Header().Set("Docker-Content-Digest", dgst.String())
	w.Header().Set("Etag", fmt.Sprintf(`"%s"`, dgst))
}

func GetManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	ref := vars["reference"]
	logrus.Info("GetManifest: ", name+":"+ref)
	mediaType := r.Header["Accept"][0]
	p, err := os.ReadFile(fileDir + ref[7:])
	if err != nil {
		return
	}
	logrus.Info("ResponseWriter@Content-length: ", len(p))
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Type", mediaType)
	w.Header().Set("Content-Length", fmt.Sprint(len(p)))
	w.Header().Set("Docker-Content-Digest", ref)
	w.Header().Set("Etag", fmt.Sprintf(`"%s"`, ref))
	w.Write(p)
}
