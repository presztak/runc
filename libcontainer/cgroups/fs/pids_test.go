// +build linux

package fs

import (
	"strconv"
	"testing"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fscommon"
)

const (
	maxUnlimited = -1
	maxLimited   = 1024
)

func TestPidsSetMax(t *testing.T) {
	helper := NewCgroupTestUtil("pids", t)

	helper.writeFileContents(map[string]string{
		"pids.max": "max",
	})

	helper.CgroupData.config.Resources.PidsLimit = maxLimited
	pids := &PidsGroup{}
	if err := pids.Set(helper.CgroupPath, helper.CgroupData.config.Resources); err != nil {
		t.Fatal(err)
	}

	value, err := fscommon.GetCgroupParamUint(helper.CgroupPath, "pids.max")
	if err != nil {
		t.Fatal(err)
	}
	if value != maxLimited {
		t.Fatalf("Expected %d, got %d for setting pids.max - limited", maxLimited, value)
	}
}

func TestPidsSetUnlimited(t *testing.T) {
	helper := NewCgroupTestUtil("pids", t)

	helper.writeFileContents(map[string]string{
		"pids.max": strconv.Itoa(maxLimited),
	})

	helper.CgroupData.config.Resources.PidsLimit = maxUnlimited
	pids := &PidsGroup{}
	if err := pids.Set(helper.CgroupPath, helper.CgroupData.config.Resources); err != nil {
		t.Fatal(err)
	}

	value, err := fscommon.GetCgroupParamString(helper.CgroupPath, "pids.max")
	if err != nil {
		t.Fatal(err)
	}
	if value != "max" {
		t.Fatalf("Expected %s, got %s for setting pids.max - unlimited", "max", value)
	}
}

func TestPidsStats(t *testing.T) {
	helper := NewCgroupTestUtil("pids", t)

	helper.writeFileContents(map[string]string{
		"pids.current": strconv.Itoa(1337),
		"pids.max":     strconv.Itoa(maxLimited),
	})

	pids := &PidsGroup{}
	stats := *cgroups.NewStats()
	if err := pids.GetStats(helper.CgroupPath, &stats); err != nil {
		t.Fatal(err)
	}

	if stats.PidsStats.Current != 1337 {
		t.Fatalf("Expected %d, got %d for pids.current", 1337, stats.PidsStats.Current)
	}

	if stats.PidsStats.Limit != maxLimited {
		t.Fatalf("Expected %d, got %d for pids.max", maxLimited, stats.PidsStats.Limit)
	}
}

func TestPidsStatsUnlimited(t *testing.T) {
	helper := NewCgroupTestUtil("pids", t)

	helper.writeFileContents(map[string]string{
		"pids.current": strconv.Itoa(4096),
		"pids.max":     "max",
	})

	pids := &PidsGroup{}
	stats := *cgroups.NewStats()
	if err := pids.GetStats(helper.CgroupPath, &stats); err != nil {
		t.Fatal(err)
	}

	if stats.PidsStats.Current != 4096 {
		t.Fatalf("Expected %d, got %d for pids.current", 4096, stats.PidsStats.Current)
	}

	if stats.PidsStats.Limit != 0 {
		t.Fatalf("Expected %d, got %d for pids.max", 0, stats.PidsStats.Limit)
	}
}
