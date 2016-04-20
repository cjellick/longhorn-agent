package controller

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/rancher/go-rancher-metadata/metadata"
	lclient "github.com/rancher/longhorn/client"
)

const (
	metadataUrl = "http://rancher-metadata/2015-12-19"
	replicaWait = 300
)

type replica struct {
	client      *metadata.Client
	host        string
	port        int
	healthState string
}

func (r replica) String() string {
	return fmt.Sprintf("tcp://%s:%d", r.host, r.port)
}

type Controller struct {
	client *lclient.ControllerClient
}

func New() *Controller {
	client := lclient.NewControllerClient("http://localhost:9501")
	return &Controller{
		client: client,
	}
}

func (c *Controller) Close() error {
	logrus.Infof("Shutting down Longhorn.")
	return nil
}

func (c *Controller) Start() error {
	logrus.Infof("Starting Longhorn")

	logrus.Info("Getting replicas.")
	replicas := []replica{}
	var scale int
	for {
		var err error
		if scale, replicas, err = c.replicas(true); err != nil {
			return err
		} else if len(replicas) < scale {
			logrus.Infof("Waiting for replicas")
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	replAddrs := make([]string, len(replicas))
	for i, r := range replicas {
		replAddrs[i] = r.String()
	}

	logrus.Infof("Starting controller with replicas: %v", replAddrs)
	if err := c.client.Start(replAddrs...); err != nil {
		return fmt.Errorf("Error starting controller: %v", err)
	}

	return c.refresh()
}

func (c *Controller) refresh() error {
	for {
		if err := c.syncReplicas(); err != nil {
			return fmt.Errorf("Failed to sync replicas: %v", err)
		}
		time.Sleep(5 * time.Second)
	}
}

func (c *Controller) syncReplicas() (retErr error) {
	logrus.Infof("Syncing replicas.")
	return nil
}

func (c *Controller) replicas(healthyOnly bool) (int, []replica, error) {
	client, err := metadata.NewClientAndWait(metadataUrl)
	if err != nil {
		return 0, nil, err
	}
	service, err := client.GetSelfServiceByName("replica")
	if err != nil {
		return 0, nil, err
	}

	result := []replica{}
	containers := map[string]metadata.Container{}
	for _, container := range service.Containers {
		if c, ok := containers[container.Name]; !ok {
			containers[container.Name] = container
		} else if container.CreateIndex > c.CreateIndex {
			containers[container.Name] = container
		}
	}

	for _, container := range containers {
		if !healthyOnly || container.HealthState == "healthy" {
			result = append(result, replica{
				healthState: container.HealthState,
				host:        container.PrimaryIp,
				port:        9502,
			})
		}
	}

	return service.Scale, result, nil
}
