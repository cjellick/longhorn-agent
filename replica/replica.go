package replica

import (
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"

	"github.com/rancher/go-rancher-metadata/metadata"
	lclient "github.com/rancher/longhorn/client"

	"github.com/rancher/longhorn-agent/controller"
)

const (
	defaultVolumeSize = "10737418240" // 10 gb
)

type Replica struct {
	client *lclient.ReplicaClient
}

func New() (*Replica, error) {
	client, err := lclient.NewReplicaClient("http://localhost:9502")
	if err != nil {
		return nil, err
	}
	return &Replica{
		client: client,
	}, nil
}

func (r *Replica) Close() error {
	logrus.Infof("Shutting down replica.")
	return nil
}

func (r *Replica) Start() error {
	_, err := metadata.NewClientAndWait(controller.MetadataURL)
	if err != nil {
		return err
	}

	// Unmarshalling the metadata as json is forcing it to a flot
	resp, err := http.Get(controller.MetadataURL + "/self/service/metadata/longhorn/volume_size")
	if err != nil {
		return err
	}

	size := ""
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		size = string(body)
	}

	if size == "" {
		size = defaultVolumeSize
	}

	logrus.Infof("Opening replica with size %v.", size)
	return r.client.OpenReplica(size)
}
