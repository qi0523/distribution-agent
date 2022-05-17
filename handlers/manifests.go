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

func GetManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	ref := vars["reference"]
	logrus.Info("GetManifest: ", name+":"+ref)
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
	if etagMatch(r, dgst.String()) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	p, err := os.ReadFile(fileDir + dgst.String()[7:])
	if err != nil {
		return
	}
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Type", mediaType)
	w.Header().Set("Content-Length", fmt.Sprint(len(p)))
	w.Header().Set("Docker-Content-Digest", dgst.String())
	w.Header().Set("Etag", fmt.Sprintf(`"%s"`, dgst))
	w.Write(p)
}

func etagMatch(r *http.Request, etag string) bool {
	for _, headerVal := range r.Header["If-None-Match"] {
		if headerVal == etag || headerVal == fmt.Sprintf(`"%s"`, etag) { // allow quoted or unquoted
			return true
		}
	}
	return false
}
