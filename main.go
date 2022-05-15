package main

import (
	"net/http"

	v2 "github.com/qi0523/distribution-agent/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	router := v2.Router()
	logrus.Info("router start....")
	if err := http.ListenAndServe(":9000", router); err != nil {
		logrus.Error("err: ", err)
	}
}
