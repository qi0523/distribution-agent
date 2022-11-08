package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	"github.com/qi0523/distribution-agent/client"
	"github.com/qi0523/distribution-agent/constant"
	"github.com/sirupsen/logrus"
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
		fi, err := os.Stat(filepath.Join(constant.ContainerdRoot, "blob/sha256", dgst.String()[7:]))
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
		p   []byte
		err error
	)

	for retry := 16; retry < 512; retry = retry << 1 {
		p, err = os.ReadFile(filepath.Join(constant.ContainerdRoot, "blobs/sha256", ref[7:]))
		if err != nil {
			time.Sleep(time.Microsecond * time.Duration(rand.Intn(retry)))
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
