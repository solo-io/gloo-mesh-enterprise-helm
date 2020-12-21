package pkg

import (
	"log"
	"os/exec"
	"runtime"
	"strings"
)

var (
	// This is the address from which pods can talk to our host machine
	DockerHostAddress = func() string {
		if runtime.GOOS == "darwin" {
			// docker for mac
			return "host.docker.internal"
		}
		// linux we need to use docker gateway ip
		ipAddr, err := exec.Command("bash", "-c", "ifconfig docker0 | grep 'inet' | cut -d: -f2 | awk '{print $1}'").CombinedOutput()
		if err != nil {
			log.Fatalf("%v", err)
		}
		ip := strings.Split(string(ipAddr), " ")[0]
		return strings.TrimSpace(ip)
	}()
)
