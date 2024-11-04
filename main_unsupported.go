// +build !linux

package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func getDefaultID() string {
	return ""
}

var (
	checkpointCommand cli.Command
	eventsCommand     cli.Command
	restoreCommand    cli.Command
	specCommand       cli.Command
	killCommand       cli.Command
)

func runAction(*cli.Context) {
	logrus.Fatal("Current OS is not supported yet")
}
