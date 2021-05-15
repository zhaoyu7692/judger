package environment

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
)

var MsPerTCK float64

func init() {
	MsPerTCK = 1.0 / float64(initialConfigHZ())
}

func B2S(bs [65]int8) string {
	var ba []byte
	for _, b := range bs {
		if b == 0 {
			break
		}
		ba = append(ba, byte(b))
	}
	return string(ba)
}

func initialConfigHZ() int {
	uname := &syscall.Utsname{}
	if err := syscall.Uname(uname); err != nil {
		panic("can't find uname")
	}
	buf := bytes.Buffer{}
	args := "/boot/config-" + B2S(uname.Release)
	cmd := exec.Command("cat", args)
	cmd.Stdout = &buf
	err := cmd.Run()
	if err == nil {
		reg := regexp.MustCompile("CONFIG_HZ=(\\d+)\n")
		result := reg.FindStringSubmatch(buf.String())
		if len(result) >= 2 {
			if TCK, err := strconv.Atoi(result[1]); err == nil && TCK > 0 {
				return TCK
			}
		}
	}
	panic("can't find TCK")
}