// +build linux

package fs

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fscommon"
)

func TestCpuSetShares(t *testing.T) {
	helper := NewCgroupTestUtil("cpu", t)

	const (
		sharesBefore = 1024
		sharesAfter  = 512
	)

	helper.writeFileContents(map[string]string{
		"cpu.shares": strconv.Itoa(sharesBefore),
	})

	helper.CgroupData.config.Resources.CpuShares = sharesAfter
	cpu := &CpuGroup{}
	if err := cpu.Set(helper.CgroupPath, helper.CgroupData.config.Resources); err != nil {
		t.Fatal(err)
	}

	value, err := fscommon.GetCgroupParamUint(helper.CgroupPath, "cpu.shares")
	if err != nil {
		t.Fatal(err)
	}
	if value != sharesAfter {
		t.Fatal("Got the wrong value, set cpu.shares failed.")
	}
}

func TestCpuSetBandWidth(t *testing.T) {
	helper := NewCgroupTestUtil("cpu", t)

	const (
		quotaBefore     = 8000
		quotaAfter      = 5000
		periodBefore    = 10000
		periodAfter     = 7000
		rtRuntimeBefore = 8000
		rtRuntimeAfter  = 5000
		rtPeriodBefore  = 10000
		rtPeriodAfter   = 7000
	)

	helper.writeFileContents(map[string]string{
		"cpu.cfs_quota_us":  strconv.Itoa(quotaBefore),
		"cpu.cfs_period_us": strconv.Itoa(periodBefore),
		"cpu.rt_runtime_us": strconv.Itoa(rtRuntimeBefore),
		"cpu.rt_period_us":  strconv.Itoa(rtPeriodBefore),
	})

	helper.CgroupData.config.Resources.CpuQuota = quotaAfter
	helper.CgroupData.config.Resources.CpuPeriod = periodAfter
	helper.CgroupData.config.Resources.CpuRtRuntime = rtRuntimeAfter
	helper.CgroupData.config.Resources.CpuRtPeriod = rtPeriodAfter
	cpu := &CpuGroup{}
	if err := cpu.Set(helper.CgroupPath, helper.CgroupData.config.Resources); err != nil {
		t.Fatal(err)
	}

	quota, err := fscommon.GetCgroupParamUint(helper.CgroupPath, "cpu.cfs_quota_us")
	if err != nil {
		t.Fatal(err)
	}
	if quota != quotaAfter {
		t.Fatal("Got the wrong value, set cpu.cfs_quota_us failed.")
	}

	period, err := fscommon.GetCgroupParamUint(helper.CgroupPath, "cpu.cfs_period_us")
	if err != nil {
		t.Fatal(err)
	}
	if period != periodAfter {
		t.Fatal("Got the wrong value, set cpu.cfs_period_us failed.")
	}

	rtRuntime, err := fscommon.GetCgroupParamUint(helper.CgroupPath, "cpu.rt_runtime_us")
	if err != nil {
		t.Fatal(err)
	}
	if rtRuntime != rtRuntimeAfter {
		t.Fatal("Got the wrong value, set cpu.rt_runtime_us failed.")
	}

	rtPeriod, err := fscommon.GetCgroupParamUint(helper.CgroupPath, "cpu.rt_period_us")
	if err != nil {
		t.Fatal(err)
	}
	if rtPeriod != rtPeriodAfter {
		t.Fatal("Got the wrong value, set cpu.rt_period_us failed.")
	}
}

func TestCpuStats(t *testing.T) {
	helper := NewCgroupTestUtil("cpu", t)

	const (
		nrPeriods     = 2000
		nrThrottled   = 200
		throttledTime = uint64(18446744073709551615)
	)

	cpuStatContent := fmt.Sprintf("nr_periods %d\nnr_throttled %d\nthrottled_time %d\n",
		nrPeriods, nrThrottled, throttledTime)
	helper.writeFileContents(map[string]string{
		"cpu.stat": cpuStatContent,
	})

	cpu := &CpuGroup{}
	actualStats := *cgroups.NewStats()
	err := cpu.GetStats(helper.CgroupPath, &actualStats)
	if err != nil {
		t.Fatal(err)
	}

	expectedStats := cgroups.ThrottlingData{
		Periods:          nrPeriods,
		ThrottledPeriods: nrThrottled,
		ThrottledTime:    throttledTime,
	}

	expectThrottlingDataEquals(t, expectedStats, actualStats.CpuStats.ThrottlingData)
}

func TestNoCpuStatFile(t *testing.T) {
	helper := NewCgroupTestUtil("cpu", t)

	cpu := &CpuGroup{}
	actualStats := *cgroups.NewStats()
	err := cpu.GetStats(helper.CgroupPath, &actualStats)
	if err != nil {
		t.Fatal("Expected not to fail, but did")
	}
}

func TestInvalidCpuStat(t *testing.T) {
	helper := NewCgroupTestUtil("cpu", t)

	cpuStatContent := `nr_periods 2000
	nr_throttled 200
	throttled_time fortytwo`
	helper.writeFileContents(map[string]string{
		"cpu.stat": cpuStatContent,
	})

	cpu := &CpuGroup{}
	actualStats := *cgroups.NewStats()
	err := cpu.GetStats(helper.CgroupPath, &actualStats)
	if err == nil {
		t.Fatal("Expected failed stat parsing.")
	}
}

func TestCpuSetRtSchedAtApply(t *testing.T) {
	helper := NewCgroupTestUtil("cpu", t)

	const (
		rtRuntimeBefore = 0
		rtRuntimeAfter  = 5000
		rtPeriodBefore  = 0
		rtPeriodAfter   = 7000
	)

	helper.writeFileContents(map[string]string{
		"cpu.rt_runtime_us": strconv.Itoa(rtRuntimeBefore),
		"cpu.rt_period_us":  strconv.Itoa(rtPeriodBefore),
	})

	helper.CgroupData.config.Resources.CpuRtRuntime = rtRuntimeAfter
	helper.CgroupData.config.Resources.CpuRtPeriod = rtPeriodAfter
	cpu := &CpuGroup{}

	helper.CgroupData.pid = 1234
	if err := cpu.Apply(helper.CgroupPath, helper.CgroupData); err != nil {
		t.Fatal(err)
	}

	rtRuntime, err := fscommon.GetCgroupParamUint(helper.CgroupPath, "cpu.rt_runtime_us")
	if err != nil {
		t.Fatal(err)
	}
	if rtRuntime != rtRuntimeAfter {
		t.Fatal("Got the wrong value, set cpu.rt_runtime_us failed.")
	}

	rtPeriod, err := fscommon.GetCgroupParamUint(helper.CgroupPath, "cpu.rt_period_us")
	if err != nil {
		t.Fatal(err)
	}
	if rtPeriod != rtPeriodAfter {
		t.Fatal("Got the wrong value, set cpu.rt_period_us failed.")
	}

	pid, err := fscommon.GetCgroupParamUint(helper.CgroupPath, "cgroup.procs")
	if err != nil {
		t.Fatal(err)
	}
	if pid != 1234 {
		t.Fatal("Got the wrong value, set cgroup.procs failed.")
	}
}
