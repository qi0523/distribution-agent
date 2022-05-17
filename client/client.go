package client

import (
	"context"
	"strings"

	"github.com/containerd/containerd"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
)

const (
	containerdPath      = "/run/containerd/containerd.sock"
	containerdNameSpace = "default"
)

var client *containerd.Client

func init() {
	var err error
	client, err = containerd.New(containerdPath, containerd.WithDefaultNamespace(containerdNameSpace))
	if err != nil {
		logrus.Fatal("failed to create containerd client.")
	}
}

func GetManifestInfoByTag(name, tag string) (string, digest.Digest, error) {
	images, err := client.ImageService().List(context.Background())
	if err != nil {
		return "", "", err
	}
	for _, image := range images {
		if image.Name[strings.Index(image.Name, "/")+1:] == name+":"+tag {
			return image.Target.MediaType, image.Target.Digest, nil
		}
	}
	return "", "", nil
}

func GetManifestInfoByDigest(dgst digest.Digest) (string, error) {
	images, err := client.ImageService().List(context.Background())
	if err != nil {
		return "", err
	}
	for _, image := range images {
		if image.Target.Digest == dgst {
			return image.Target.MediaType, nil
		}
	}
	return "", nil
}
