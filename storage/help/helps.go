package help

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/opencontainers/go-digest"
)

const (
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

func BlobPath(ref string) string {
	return "blobs/sha256/" + ref[7:]
}

func IngestPath(mediaType, ref string) string {
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
