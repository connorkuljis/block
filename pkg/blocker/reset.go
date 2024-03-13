package blocker

import (
	"os/exec"
	"runtime"
)

func ResetDNS() error {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("sudo", "dscacheutil", "-flushcache")
		err := cmd.Run()
		if err != nil {
			return err
		}

		cmd = exec.Command("sudo", "killall", "-HUP", "mDNSResponder")
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
