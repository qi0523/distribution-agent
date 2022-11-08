package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	"github.com/qi0523/distribution-agent/client"
	"github.com/sirupsen/logrus"
)

const (
	fileDir = "/mydata/var/lib/containerd/io.containerd.content.v1.content/blobs/sha256/"
)

func ResolvedManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	ref := vars["reference"]
	logrus.Info("ResolvedManifest: ", name+":"+ref)
	dgst, err := digest.Parse(ref)
	var (
		mediaType string
		size      int64
	)
	if err != nil { //tag
		mediaType, dgst, err = client.GetManifestInfoByTag(name, ref)
	} else {
		mediaType, err = client.GetManifestInfoByDigest(dgst)
	}
	if err != nil || mediaType == "" {
		if mediaType, dgst, size, err = client.GetManifestInfoByTmpImage(name, ref); err != nil {
			return
		}
	} else {
		fi, err := os.Stat(fileDir + dgst.String()[7:])
		if err != nil {
			return
		}
		size = fi.Size()
	}
	logrus.Info("ResponseWriter@Content-length: ", size)
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Type", mediaType)
	w.Header().Set("Content-Length", fmt.Sprint(size))
	w.Header().Set("Docker-Content-Digest", dgst.String())
	w.Header().Set("Etag", fmt.Sprintf(`"%s"`, dgst))
}

func GetManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	ref := vars["reference"]
	logrus.Info("GetManifest: ", name+":"+ref)
	mediaType := r.Header["Accept"][0]
	var (
		p     []byte
		err   error
		retry = 16
	)

	for {
		p, err = os.ReadFile(fileDir + ref[7:])
		if err != nil {
			if retry < 512 {
				time.Sleep(time.Microsecond * time.Duration(rand.Intn(retry)))
				retry = retry << 1
			} else {
				return
			}
		} else {
			break
		}
	}
	logrus.Info("ResponseWriter@Content-length: ", len(p))
	w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Type", mediaType)
	w.Header().Set("Content-Length", fmt.Sprint(len(p)))
	w.Header().Set("Docker-Content-Digest", ref)
	w.Header().Set("Etag", fmt.Sprintf(`"%s"`, ref))
	w.Write(p)
}
