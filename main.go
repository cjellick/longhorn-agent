package main

import (
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
	"github.com/rancher/longhorn-agent/controller"
	"github.com/rancher/longhorn-agent/replica"
	"github.com/rancher/longhorn-agent/status"
)

var (
	VERSION = "0.0.0"
)

func main() {
	app := cli.NewApp()
	app.Version = VERSION
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "controller",
		},
		cli.BoolFlag{
			Name: "replica",
		},
		cli.StringFlag{
			Name:  "listen, l",
			Value: ":8199",
		},
		cli.StringFlag{
			Name:  "log-level",
			Value: "debug",
		},
	}
	app.Action = func(c *cli.Context) {
		if err := runApp(c); err != nil {
			logrus.Fatal(err)
		}
	}
	app.Run(os.Args)
}

func runApp(context *cli.Context) error {
	logLevel := context.GlobalString("log-level")
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)

	runController := context.GlobalBool("controller")
	runReplica := context.GlobalBool("replica")

	if runController {
		go runPing(context)
		c := controller.New()
		defer c.Close()
		return c.Start()
	} else if runReplica {
		// TODO Refactor to not suck. Put start in same container as actual repl process?
		r, err := replica.New()
		if err != nil {
			return err
		}
		defer r.Close()
		if err := r.Start(); err != nil {
			return err
		}
		return runPing(context)
	}

	return nil
}

func runPing(context *cli.Context) error {
	status, err := status.New()
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/controller/status", status.ControllerStatus)
	r.HandleFunc("/replica/status", status.ReplicaStatus)
	http.Handle("/", r)

	listen := context.GlobalString("listen")
	logrus.Info("Listening on ", listen)
	return http.ListenAndServe(listen, nil)
}