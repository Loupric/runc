// +build linux

package fs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Loupric/runc/cgroups"
	"github.com/Loupric/runc/configs"
)

type HugetlbGroup struct {
}

func (s *HugetlbGroup) Apply(d *data) error {
	dir, err := d.join("hugetlb")
	if err != nil && !cgroups.IsNotFound(err) {
		return err
	}

	if err := s.Set(dir, d.c); err != nil {
		return err
	}

	return nil
}

func (s *HugetlbGroup) Set(path string, cgroup *configs.Cgroup) error {
	for _, hugetlb := range cgroup.HugetlbLimit {
		if err := writeFile(path, strings.Join([]string{"hugetlb", hugetlb.Pagesize, "limit_in_bytes"}, "."), strconv.Itoa(hugetlb.Limit)); err != nil {
			return err
		}
	}

	return nil
}

func (s *HugetlbGroup) Remove(d *data) error {
	return removePath(d.path("hugetlb"))
}

func (s *HugetlbGroup) GetStats(path string, stats *cgroups.Stats) error {
	hugetlbStats := cgroups.HugetlbStats{}
	for _, pageSize := range HugePageSizes {
		usage := strings.Join([]string{"hugetlb", pageSize, "usage_in_bytes"}, ".")
		value, err := getCgroupParamUint(path, usage)
		if err != nil {
			return fmt.Errorf("failed to parse %s - %v", usage, err)
		}
		hugetlbStats.Usage = value

		maxUsage := strings.Join([]string{"hugetlb", pageSize, "max_usage_in_bytes"}, ".")
		value, err = getCgroupParamUint(path, maxUsage)
		if err != nil {
			return fmt.Errorf("failed to parse %s - %v", maxUsage, err)
		}
		hugetlbStats.MaxUsage = value

		failcnt := strings.Join([]string{"hugetlb", pageSize, "failcnt"}, ".")
		value, err = getCgroupParamUint(path, failcnt)
		if err != nil {
			return fmt.Errorf("failed to parse %s - %v", failcnt, err)
		}
		hugetlbStats.Failcnt = value

		stats.HugetlbStats[pageSize] = hugetlbStats
	}

	return nil
}
