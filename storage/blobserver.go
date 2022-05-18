package storage

import (
	"fmt"
	"net/http"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/qi0523/distribution-agent/storage/driver"
)

const blobCacheControlMaxAge = 365 * 24 * time.Hour

type blobServer struct {
	driver driver.StorageDriver
}

func (bs *blobServer) ServeBlob(w http.ResponseWriter, r *http.Request, dgst digest.Digest) error {
	size, err := bs.driver.Stat(dgst.String())
	if err != nil {
		return err
	}
	br, err := newFileReader(bs.driver, dgst.String(), size)
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
	if w.Header().Get("Content-Length") == "" {
		// Set the content length if not already set.
		w.Header().Set("Content-Length", fmt.Sprint(size))
	}
	http.ServeContent(w, r, dgst.String(), time.Time{}, br)
	return nil
}
