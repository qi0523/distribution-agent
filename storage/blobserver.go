package storage

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "crypto/sha256"

	digest "github.com/opencontainers/go-digest"
	"github.com/qi0523/distribution-agent/constant"
	"github.com/qi0523/distribution-agent/storage/driver"
	"github.com/qi0523/distribution-agent/storage/driver/factory"
	"github.com/qi0523/distribution-agent/storage/help"

	_ "github.com/qi0523/distribution-agent/storage/driver/filesystem"
)

const (
	defaultDriverName      = "filesystem"
	blobCacheControlMaxAge = 365 * 24 * time.Hour
)

type blobServer struct {
	driver driver.StorageDriver
}

var bs *blobServer
var once sync.Once

func GetBlobServer() *blobServer {
	once.Do(func() {
		d, err := factory.Create(defaultDriverName, nil)
		if err != nil {
			panic(err)
		}
		bs = &blobServer{
			driver: d,
		}
	})
	return bs
}

func (bs *blobServer) ServeBlob(w http.ResponseWriter, r *http.Request, dgst digest.Digest) error {
	var (
		err       error
		size      int64
		br        *fileReader
		path      string
		mediaType = strings.Split(r.Header["Accept"][0], ",")[0]
	)
	ch := make(chan bool, 1)
	go func() {
		for {
			path = help.BlobPath(dgst.String())
			if size, err = bs.driver.Stat(path); err == nil {
				if br, err = newFileReader(bs.driver, dgst.String(), mediaType, size); err == nil {
					ch <- true
					break
				}
			} else {
				var totalS string
				path = help.IngestPath(mediaType, dgst.String())
				if totalS, err = readFileString(filepath.Join(constant.ContainerdRoot, path, "total")); err == nil {
					if size, err = strconv.ParseInt(totalS, 10, 64); err == nil {
						if br, err = newFileReader(bs.driver, dgst.String(), mediaType, size); err == nil {
							ch <- true
							break
						}
					}
				}
			}
			time.Sleep(time.Millisecond * time.Duration(constant.Interval))
		}
	}()
	select {
	case _ = <-ch:
		defer br.Close()
		w.Header().Set("ETag", fmt.Sprintf(`"%s`, dgst))
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%.f", blobCacheControlMaxAge.Seconds()))
		if w.Header().Get("Docker-Content-Digest") == "" {
			w.Header().Set("Docker-Content-Digest", dgst.String())
		}
		//MediaType ?
		if w.Header().Get("Content-Type") == "" {
			// Set the content type if not already set.
			w.Header().Set("Content-Type", r.Header["Accept"][0])
		}

		if w.Header().Get("Content-Length") == "" {
			// Set the content length if not already set.
			w.Header().Set("Content-Length", fmt.Sprint(size))
		}
		http.ServeContent(w, r, dgst.String(), time.Time{}, br)
		return nil
	case <-time.After(constant.ServeBlobTimeout * time.Second):
		w.WriteHeader(http.StatusNotFound)
		return err
	}
}

func readFileString(path string) (string, error) {
	p, err := os.ReadFile(path)
	return string(p), err
}
