package storage

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/qi0523/distribution-agent/storage/driver"
	"github.com/qi0523/distribution-agent/storage/driver/factory"
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
	size, err := bs.driver.Stat(dgst.String()[7:])
	if err != nil {
		return err
	}
	br, err := newFileReader(bs.driver, dgst.String()[7:], size)
	if err != nil {
		return err
	}
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
}
