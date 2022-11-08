package client

import (
	"context"
	"strings"

	"github.com/containerd/containerd"
	"github.com/opencontainers/go-digest"
	"github.com/qi0523/distribution-agent/constant"
	"github.com/sirupsen/logrus"
)

var client *containerd.Client

func init() {
	var err error
	client, err = containerd.New(constant.ContainerdSockPath, containerd.WithDefaultNamespace(constant.ContainerdNameSpace))
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

func GetManifestInfoByTmpImage(name, tag string) (string, digest.Digest, int64, error) {
	tmpimage, err := client.GetTmpImage(context.Background(), name+":"+tag)
	if err != nil {
		return "", "", 0, err
	}
	return tmpimage.Target.MediaType, tmpimage.Target.Digest, tmpimage.Target.Size, nil
}
