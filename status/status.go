package status

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

type Client struct {
}

func New() (*Client, error) {
	return &Client{}, nil
}

func (s *Client) ControllerStatus(rw http.ResponseWriter, r *http.Request) {
	writeOk(rw)
}

func (s *Client) ReplicaStatus(rw http.ResponseWriter, r *http.Request) {
	writeOk(rw)
}

func writeOk(rw http.ResponseWriter) {
	logrus.Debugf("Reporting OK.")
	rw.Write([]byte("OK"))
}

func writeError(rw http.ResponseWriter, err error) {
	writeErrorString(rw, err.Error())
}

func writeErrorString(rw http.ResponseWriter, msg string) {
	if rw != nil {
		logrus.Infof("Reporting unhealthy status: %v", msg)
		rw.WriteHeader(http.StatusServiceUnavailable)
		rw.Write([]byte(msg))
	}
}
