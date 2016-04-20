package replica

import (
	"github.com/Sirupsen/logrus"

	lclient "github.com/rancher/longhorn/client"
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

func (c *Replica) Start() error {
	logrus.Infof("Opening replica.")
	err := c.client.OpenReplica("10737418240") // hard code 10 gigs
	return err
}



