package status

import (
	"fmt"
	"net/http"

	md "github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/longhorn/client"

	"github.com/rancher/longhorn-agent/controller"
	"strings"
)

type ReplicaStatus struct {
	controller *client.ControllerClient
	replica    *client.ReplicaClient
	metadata   *md.Client
	address    string
}

func NewReplicaStatus() (*ReplicaStatus, error) {
	metadata, err := md.NewClientAndWait(controller.MetadataURL)
	if err != nil {
		return nil, err
	}
	self, err := metadata.GetSelfContainer()
	if err != nil {
		return nil, err
	}
	addr := controller.ReplicaAddress(self.PrimaryIp, 9502)

	controllerClient := client.NewControllerClient("http://controller:9501/v1")
	replicaClient, err := client.NewReplicaClient("http://localhost:9502/v1")
	if err != nil {
		return nil, err
	}

	return &ReplicaStatus{
		controller: controllerClient,
		replica:    replicaClient,
		address:    addr,
	}, nil
}

func (s *ReplicaStatus) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// Checking against the replica is easy: just ensure that the API is responding.
	_, err := s.replica.GetReplica()
	if err != nil {
		writeError(rw, err)
		return
	}

	if err := s.checkReplicaStatusInController(rw); err != nil {
		writeError(rw, err)
		return
	}

	writeOK(rw)
}

func (s *ReplicaStatus) checkReplicaStatusInController(rw http.ResponseWriter) error {
	replicas, err := s.controller.ListReplicas()
	if err != nil {
		// TODO If this errors out, we should probably return healthy as a cached response
		return fmt.Errorf("Couldn't get replicas from controller. Error: %v", err)
	}
	for _, replica := range replicas {
		if replica.Address == s.address {
			if strings.EqualFold(replica.Mode, "err") {
				return fmt.Errorf("Replica %v is in error mode.", s.address)
			}
			return nil // Replica is healthy
		}
	}

	return fmt.Errorf("Replica %v is not in the controller's list of replicas. Current list: %v", s.address, replicas)
}
