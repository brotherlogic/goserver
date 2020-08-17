package goserver

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func (s *GoServer) hasScratch() bool {
	file, err := os.Open("/proc/mounts")

	if err != nil {
		s.Log(fmt.Sprintf("Unable to read mounts: %v", err))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 2 && fields[1] == "/media/scratch" {
			return true
		}
	}

	return false

}

func (s *GoServer) prepDLog() {
	if s.hasScratch() {
		s.Log(fmt.Sprintf("Prepping for Disk Logging"))
	} else {
		s.Log(fmt.Sprintf("Scratch not found, no disk logging"))
	}
}
