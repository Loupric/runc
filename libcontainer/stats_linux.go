package libcontainer

import "github.com/Loupric/runc/cgroups"

type Stats struct {
	Interfaces  []*NetworkInterface
	CgroupStats *cgroups.Stats
}
