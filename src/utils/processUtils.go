package utils

import (
	"environment"
	"io/ioutil"
	"math"
	"regexp"
	"strconv"
)
// ms
func GetTimeUsed(pid int) int {
	if content, err := ioutil.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat"); err == nil {
		reg := regexp.MustCompile("\\d+ \\(.*\\) \\w \\d+ \\d+ \\d+ \\d+ \\d+ \\d+ \\d+ \\d+ \\d+ \\d+ (\\d+) (\\d+)")
		result := reg.FindStringSubmatch(string(content))
		userTime, _ := strconv.Atoi(result[1])
		kernelTime, _ := strconv.Atoi(result[2])
		return int(math.Floor(float64(userTime+kernelTime) * environment.MsPerTCK * 1000))
	}
	return -1
}
// KBs
func GetMemoryUsed(pid int) int {
	if content, err := ioutil.ReadFile("/proc/" + strconv.Itoa(pid) + "/statm"); err == nil {
		reg := regexp.MustCompile("\\d+ (\\d+) \\d+ \\d+ \\d+ \\d+ \\d+")
		result := reg.FindStringSubmatch(string(content))
		memoryUsed, _ := strconv.Atoi(result[1])
		return memoryUsed * 4
	}
	return -1
}

func GetWriteBytes(pid int) int  {
	if content, err := ioutil.ReadFile("/proc/" + strconv.Itoa(pid) + "/io"); err == nil {
		reg := regexp.MustCompile("wchar: (\\d+)")
		result := reg.FindStringSubmatch(string(content))
		writeBytes, _ := strconv.Atoi(result[1])
		return writeBytes
	}
	return -1
}