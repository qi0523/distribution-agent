package handlers

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func GetBlobs(w http.ResponseWriter, r *http.Request) {
	logrus.Info("ctx: ", r.Context())
}
