// +build linux

package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/Loupric/runc"
	"github.com/opencontainers/runtime-spec/specs-go"
)

const SD_LISTEN_FDS_START = 3

// default action is to start a container
var startCommand = cli.Command{
	Name:  "start",
	Usage: "create and run a container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "config-file, c",
			Value: "config.json",
			Usage: "path to spec config file",
		},
		cli.StringFlag{
			Name:  "runtime-file, r",
			Value: "runtime.json",
			Usage: "path to runtime config file",
		},
	},
	Action: func(context *cli.Context) {
		spec, rspec, err := loadSpec(context.String("config-file"), context.String("runtime-file"))
		if err != nil {
			fatal(err)
		}

		notifySocket := os.Getenv("NOTIFY_SOCKET")
		if notifySocket != "" {
			setupSdNotify(spec, rspec, notifySocket)
		}

		listenFds := os.Getenv("LISTEN_FDS")
		listenPid := os.Getenv("LISTEN_PID")

		if listenFds != "" && listenPid == strconv.Itoa(os.Getpid()) {
			setupSocketActivation(spec, listenFds)
		}

		if os.Geteuid() != 0 {
			logrus.Fatal("runc should be run as root")
		}
		status, err := startContainer(context, spec, rspec)
		if err != nil {
			logrus.Fatalf("Container start failed: %v", err)
		}
		// exit with the container's exit status so any external supervisor is
		// notified of the exit with the correct exit status.
		os.Exit(status)
	},
}

func init() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
		factory, _ := libcontainer.New("")
		if err := factory.StartInitialization(); err != nil {
			fatal(err)
		}
		panic("--this line should have never been executed, congratulations--")
	}
}

func startContainer(context *cli.Context, spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec) (int, error) {
	config, err := createLibcontainerConfig(context.GlobalString("id"), spec, rspec)
	if err != nil {
		return -1, err
	}
	if _, err := os.Stat(config.Rootfs); err != nil {
		if os.IsNotExist(err) {
			return -1, fmt.Errorf("Rootfs (%q) does not exist", config.Rootfs)
		}
		return -1, err
	}
	rootuid, err := config.HostUID()
	if err != nil {
		return -1, err
	}
	factory, err := loadFactory(context)
	if err != nil {
		return -1, err
	}
	container, err := factory.Create(context.GlobalString("id"), config)
	if err != nil {
		return -1, err
	}
	// ensure that the container is always removed if we were the process
	// that created it.
	defer destroy(container)
	process := newProcess(spec.Process)

	// Support on-demand socket activation by passing file descriptors into the container init process.
	if os.Getenv("LISTEN_FDS") != "" {
		listenFdsInt, err := strconv.Atoi(os.Getenv("LISTEN_FDS"))
		if err != nil {
			return -1, err
		}

		for i := SD_LISTEN_FDS_START; i < (listenFdsInt + SD_LISTEN_FDS_START); i++ {
			process.ExtraFiles = append(process.ExtraFiles, os.NewFile(uintptr(i), ""))
		}
	}

	tty, err := newTty(spec.Process.Terminal, process, rootuid)
	if err != nil {
		return -1, err
	}
	handler := newSignalHandler(tty)
	defer handler.Close()
	if err := container.Start(process); err != nil {
		return -1, err
	}
	return handler.forward(process)
}

// If systemd is supporting sd_notify protocol, this function will add support
// for sd_notify protocol from within the container.
func setupSdNotify(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, notifySocket string) {
	mountName := "sdNotify"
	spec.Mounts = append(spec.Mounts, specs.MountPoint{Name: mountName, Path: notifySocket})
	spec.Process.Env = append(spec.Process.Env, fmt.Sprintf("NOTIFY_SOCKET=%s", notifySocket))
	rspec.Mounts[mountName] = specs.Mount{Type: "bind", Source: notifySocket, Options: []string{"bind"}}
}

// If systemd is supporting on-demand socket activation, this function will add support
// for on-demand socket activation for the containerized service.
func setupSocketActivation(spec *specs.LinuxSpec, listenFds string) {
	spec.Process.Env = append(spec.Process.Env, fmt.Sprintf("LISTEN_FDS=%s", listenFds), "LISTEN_PID=1")
}

func destroy(container libcontainer.Container) {
	status, err := container.Status()
	if err != nil {
		logrus.Error(err)
	}
	if status != libcontainer.Checkpointed {
		if err := container.Destroy(); err != nil {
			logrus.Error(err)
		}
	}
}
