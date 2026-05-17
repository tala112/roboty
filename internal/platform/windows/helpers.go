//go:build windows

package windows

import (
	"fmt"
	"os/exec"
)

func fmtErr(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func execCmd(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}
