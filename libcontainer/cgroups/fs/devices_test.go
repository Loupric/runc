// +build linux

package fs

import (
	"testing"

	"github.com/Loupric/runc/configs"
)

var (
	allowedDevices = []*configs.Device{
		{
			Path:        "/dev/zero",
			Type:        'c',
			Major:       1,
			Minor:       5,
			Permissions: "rwm",
			FileMode:    0666,
		},
	}
	allowedList   = "c 1:5 rwm"
	deniedDevices = []*configs.Device{
		{
			Path:        "/dev/null",
			Type:        'c',
			Major:       1,
			Minor:       3,
			Permissions: "rwm",
			FileMode:    0666,
		},
	}
	deniedList = "c 1:3 rwm"
)

func TestDevicesSetAllow(t *testing.T) {
	helper := NewCgroupTestUtil("devices", t)
	defer helper.cleanup()

	helper.writeFileContents(map[string]string{
		"devices.deny": "a",
	})

	helper.CgroupData.c.AllowAllDevices = false
	helper.CgroupData.c.AllowedDevices = allowedDevices
	devices := &DevicesGroup{}
	if err := devices.Set(helper.CgroupPath, helper.CgroupData.c); err != nil {
		t.Fatal(err)
	}

	value, err := getCgroupParamString(helper.CgroupPath, "devices.allow")
	if err != nil {
		t.Fatalf("Failed to parse devices.allow - %s", err)
	}

	if value != allowedList {
		t.Fatal("Got the wrong value, set devices.allow failed.")
	}
}

func TestDevicesSetDeny(t *testing.T) {
	helper := NewCgroupTestUtil("devices", t)
	defer helper.cleanup()

	helper.writeFileContents(map[string]string{
		"devices.allow": "a",
	})

	helper.CgroupData.c.AllowAllDevices = true
	helper.CgroupData.c.DeniedDevices = deniedDevices
	devices := &DevicesGroup{}
	if err := devices.Set(helper.CgroupPath, helper.CgroupData.c); err != nil {
		t.Fatal(err)
	}

	value, err := getCgroupParamString(helper.CgroupPath, "devices.deny")
	if err != nil {
		t.Fatalf("Failed to parse devices.deny - %s", err)
	}

	if value != deniedList {
		t.Fatal("Got the wrong value, set devices.deny failed.")
	}
}
