package goserver

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
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
		filename := fmt.Sprintf("/media/scratch/dlogs/%v/%v.logs", s.Registry.GetName(), time.Now().Unix())
		os.MkdirAll(fmt.Sprintf("/media/scratch/dlogs/%v-dlogs/", s.Registry.GetName()), 0777)
		fhandle, err := os.Create(filename)
		if err != nil {
			s.Log(fmt.Sprintf("Unable to open file: %v", err))
		}
		s.dlogHandle = fhandle
	} else {
		s.Log(fmt.Sprintf("Scratch not found, no disk logging"))
	}
}
