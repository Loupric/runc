// +build linux

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli"
	"github.com/Loupric/runc"
)

var checkpointCommand = cli.Command{
	Name:  "checkpoint",
	Usage: "checkpoint a running container",
	Flags: []cli.Flag{
		cli.StringFlag{Name: "image-path", Value: "", Usage: "path for saving criu image files"},
		cli.StringFlag{Name: "work-path", Value: "", Usage: "path for saving work files and logs"},
		cli.BoolFlag{Name: "leave-running", Usage: "leave the process running after checkpointing"},
		cli.BoolFlag{Name: "tcp-established", Usage: "allow open tcp connections"},
		cli.BoolFlag{Name: "ext-unix-sk", Usage: "allow external unix sockets"},
		cli.BoolFlag{Name: "shell-job", Usage: "allow shell jobs"},
		cli.StringFlag{Name: "page-server", Value: "", Usage: "ADDRESS:PORT of the page server"},
		cli.BoolFlag{Name: "file-locks", Usage: "handle file locks, for safety"},
		cli.BoolFlag{Name: "pre-dump", Usage: "Make pre dump"},
		cli.StringFlag{Name: "prev-images-dir", Usage: "Previous img dir"},
		cli.BoolFlag{Name: "track-mem", Usage: "Turn on memory changes tracker in kernel"},
	},
	Action: func(context *cli.Context) {
		container, err := getContainer(context)
		if err != nil {
			fatal(err)
		}
		options := criuOptions(context)
		status, err := container.Status()
		if err != nil {
			fatal(err)
		}
		if status == libcontainer.Checkpointed {
			fatal(fmt.Errorf("Container with id %s already checkpointed", context.GlobalString("id")))
		}
		// these are the mandatory criu options for a container
		setPageServer(context, options)
		if err := container.Checkpoint(options); err != nil {
			fatal(err)
		}
	},
}

func getCheckpointImagePath(context *cli.Context) string {
	imagePath := context.String("image-path")
	if imagePath == "" {
		imagePath = getDefaultImagePath(context)
	}
	return imagePath
}

func setPageServer(context *cli.Context, options *libcontainer.CriuOpts) {
	// xxx following criu opts are optional
	// The dump image can be sent to a criu page server
	if psOpt := context.String("page-server"); psOpt != "" {
		addressPort := strings.Split(psOpt, ":")
		if len(addressPort) != 2 {
			fatal(fmt.Errorf("Use --page-server ADDRESS:PORT to specify page server"))
		}
		portInt, err := strconv.Atoi(addressPort[1])
		if err != nil {
			fatal(fmt.Errorf("Invalid port number"))
		}
		options.PageServer = libcontainer.CriuPageServerInfo{
			Address: addressPort[0],
			Port:    int32(portInt),
		}
	}
}
