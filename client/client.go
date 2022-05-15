package client

import (
	"sync"

	"github.com/containerd/containerd"
	"github.com/sirupsen/logrus"
)

const (
	containerdPath      = "/run/containerd/containerd.sock"
	containerdNameSpace = "default"
)

var client *containerd.Client
var once sync.Once

//GetInstance
func GetInstance() *containerd.Client {
	once.Do(func() {
		var err error
		client, err = containerd.New(containerdPath, containerd.WithDefaultNamespace(containerdNameSpace))
		if err != nil {
			logrus.Fatal("failed to create containerd client.")
		}
	})
	return client
}
