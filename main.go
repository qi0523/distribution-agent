package main

import (
	"net/http"
	"os"

	v2 "github.com/qi0523/distribution-agent/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	router := v2.Router()
	listenPort := os.Args[1]
	logrus.Info("router start....")
	if err := http.ListenAndServe(":"+listenPort, router); err != nil {
		logrus.Error("err: ", err)
	}
}
