package handlers

import (
	"fmt"
	"io/fs"
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
	var (
		mediaType string
		size      int64
		dgst      digest.Digest
		fi        fs.FileInfo
		err       error
	)
	ch := make(chan bool, 1)
	go func() {
		for {
			if mediaType, dgst, size, err = client.GetManifestInfoByTmpImage(name, ref); err == nil {
				ch <- true
				break
			} else {
				if mediaType, dgst, err = client.GetManifestInfoByTag(name, ref); mediaType != "" {
					if fi, err = os.Stat(filepath.Join(constant.ContainerdRoot, "blobs/sha256", dgst.String()[7:])); err == nil {
						size = fi.Size()
						ch <- true
						break
					}
				}
			}
			time.Sleep(time.Millisecond * time.Duration(constant.Interval))
		}
	}()

	select {
	case _ = <-ch:
		logrus.Info("ResponseWriter@Content-length: ", size)
		w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
		w.Header().Set("Content-Type", mediaType)
		w.Header().Set("Content-Length", fmt.Sprint(size))
		w.Header().Set("Docker-Content-Digest", dgst.String())
		w.Header().Set("Etag", fmt.Sprintf(`"%s"`, dgst))
	case <-time.After(constant.ResolvedTimeout * time.Second):
		w.WriteHeader(http.StatusNotFound)
		return
	}
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

	ch := make(chan bool, 1)
	go func() {
		for {
			if p, err = os.ReadFile(filepath.Join(constant.ContainerdRoot, "blobs/sha256", ref[7:])); err == nil {
				ch <- true
				break
			}
			time.Sleep(time.Millisecond * time.Duration(constant.Interval))
		}
	}()

	select {
	case _ = <-ch:
		logrus.Info("ResponseWriter@Content-length: ", len(p))
		w.Header().Add("Docker-Distribution-API-Version", "registry/2.0")
		w.Header().Set("Content-Type", mediaType)
		w.Header().Set("Content-Length", fmt.Sprint(len(p)))
		w.Header().Set("Docker-Content-Digest", ref)
		w.Header().Set("Etag", fmt.Sprintf(`"%s"`, ref))
		w.Write(p)
	case <-time.After(constant.ResolvedTimeout * time.Second):
		w.WriteHeader(http.StatusNotFound)
		return
	}
}
