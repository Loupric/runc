// +build linux

package fs

import (
	"github.com/Loupric/runc/cgroups"
	"github.com/Loupric/runc/configs"
)

type DevicesGroup struct {
}

func (s *DevicesGroup) Apply(d *data) error {
	dir, err := d.join("devices")
	if err != nil {
		// We will return error even it's `not found` error, devices
		// cgroup is hard requirement for container's security.
		return err
	}

	if err := s.Set(dir, d.c); err != nil {
		return err
	}

	return nil
}

func (s *DevicesGroup) Set(path string, cgroup *configs.Cgroup) error {
	if !cgroup.AllowAllDevices {
		if err := writeFile(path, "devices.deny", "a"); err != nil {
			return err
		}

		for _, dev := range cgroup.AllowedDevices {
			if err := writeFile(path, "devices.allow", dev.CgroupString()); err != nil {
				return err
			}
		}
		return nil
	}

	if err := writeFile(path, "devices.allow", "a"); err != nil {
		return err
	}

	for _, dev := range cgroup.DeniedDevices {
		if err := writeFile(path, "devices.deny", dev.CgroupString()); err != nil {
			return err
		}
	}

	return nil
}

func (s *DevicesGroup) Remove(d *data) error {
	return removePath(d.path("devices"))
}

func (s *DevicesGroup) GetStats(path string, stats *cgroups.Stats) error {
	return nil
}
