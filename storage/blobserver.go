package storage

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "crypto/sha256"

	digest "github.com/opencontainers/go-digest"
	"github.com/qi0523/distribution-agent/constant"
	"github.com/qi0523/distribution-agent/storage/driver"
	"github.com/qi0523/distribution-agent/storage/driver/factory"
	_ "github.com/qi0523/distribution-agent/storage/driver/filesystem"
)

const (
	defaultDriverName                      = "filesystem"
	blobCacheControlMaxAge                 = 365 * 24 * time.Hour
	MediaTypeDockerSchema2Layer            = "application/vnd.docker.image.rootfs.diff.tar"
	MediaTypeDockerSchema2LayerForeign     = "application/vnd.docker.image.rootfs.foreign.diff.tar"
	MediaTypeDockerSchema2LayerGzip        = "application/vnd.docker.image.rootfs.diff.tar.gzip"
	MediaTypeDockerSchema2LayerForeignGzip = "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip"
	MediaTypeDockerSchema2Manifest         = "application/vnd.docker.distribution.manifest.v2+json"
	MediaTypeDockerSchema2ManifestList     = "application/vnd.docker.distribution.manifest.list.v2+json"
	MediaTypeImageManifest                 = "application/vnd.oci.image.manifest.v1+json"
	MediaTypeImageIndex                    = "application/vnd.oci.image.index.v1+json"
	MediaTypeImageConfig                   = "application/vnd.oci.image.config.v1+json"
	MediaTypeDockerSchema2Config           = "application/vnd.docker.container.image.v1+json"
	MediaTypeContainerd1Checkpoint         = "application/vnd.containerd.container.criu.checkpoint.criu.tar"
	MediaTypeContainerd1CheckpointConfig   = "application/vnd.containerd.container.checkpoint.config.v1+proto"
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
		mediaType = r.Header["Accept"][0]
		retry     = 16
	)
	path = blobPath(dgst.String())
	size, err = bs.driver.Stat(path)
	if err == nil {
		br, err = newFileReader(bs.driver, path, size)
		defer br.Close()
		if err != nil {
			return err
		}
	} else {
		var totalS string
		path = ingestPath(mediaType, dgst.String())
		for {
			totalS, err = readFileString(filepath.Join(constant.ContainerdRoot, path, "total"))
			if err != nil {
				if retry < 512 {
					time.Sleep(time.Microsecond * time.Duration(rand.Intn(retry)))
					retry = retry << 1
				} else {
					return err
				}
			} else {
				break
			}
		}
		if size, err = strconv.ParseInt(totalS, 10, 64); err != nil {
			return err
		}
		br, err = newFileReader(bs.driver, filepath.Join(path, "data"), size)
		defer br.Close()
		if err != nil {
			return err
		}
	}
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

func blobPath(ref string) string {
	return "blobs/sha256/" + ref[7:]
}

func ingestPath(mediaType, ref string) string {
	key := makeRefKey(mediaType, ref)
	dgst := digest.FromString("default/1/" + key)
	return filepath.Join("ingest", dgst.Hex())
}

func makeRefKey(mediaType, key string) string {
	switch mt := mediaType; {
	case mt == MediaTypeDockerSchema2Manifest || mt == MediaTypeImageManifest:
		return "manifest-" + key
	case mt == MediaTypeDockerSchema2ManifestList || mt == MediaTypeImageIndex:
		return "index-" + key
	case isLayerType(mt):
		return "layer-" + key
	case isKnownConfig(mt):
		return "config-" + key
	default:
		return "unknown-" + key
	}
}

func parseMediaTypes(mt string) (string, []string) {
	if mt == "" {
		return "", []string{}
	}

	s := strings.Split(mt, "+")
	ext := s[1:]
	sort.Strings(ext)

	return s[0], ext
}

// IsLayerType returns true if the media type is a layer
func isLayerType(mt string) bool {
	if strings.HasPrefix(mt, "application/vnd.oci.image.layer.") {
		return true
	}

	// Parse Docker media types, strip off any + suffixes first
	base, _ := parseMediaTypes(mt)
	switch base {
	case MediaTypeDockerSchema2Layer, MediaTypeDockerSchema2LayerGzip,
		MediaTypeDockerSchema2LayerForeign, MediaTypeDockerSchema2LayerForeignGzip:
		return true
	}
	return false
}

func isKnownConfig(mt string) bool {
	switch mt {
	case MediaTypeDockerSchema2Config, MediaTypeImageConfig,
		MediaTypeContainerd1Checkpoint, MediaTypeContainerd1CheckpointConfig:
		return true
	}
	return false
}

func readFileString(path string) (string, error) {
	p, err := os.ReadFile(path)
	return string(p), err
}
