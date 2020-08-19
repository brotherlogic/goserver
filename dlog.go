package goserver

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

var (
	logSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "dlog_size",
		Help: "The number of server requests",
	})
)

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func (s *GoServer) DLog(text string) {
	if s.dlogHandle != nil {
		s.dlogHandle.WriteString(fmt.Sprintf("%v %v\n", time.Now(), text))

		size, err := dirSize(fmt.Sprintf("/media/scratch/dlogs/%v", s.Registry.GetName()))
		if err != nil {
			s.RaiseIssue("Bad log problem", fmt.Sprintf("Error reeading logs: %v", err))
		} else {
			logSize.Set(float64(size))
		}

	}
}

func (s *GoServer) prepDLog() {
	if s.hasScratch() {
		filename := fmt.Sprintf("/media/scratch/dlogs/%v/%v.logs", s.Registry.GetName(), time.Now().Unix())
		os.MkdirAll(fmt.Sprintf("/media/scratch/dlogs/%v/", s.Registry.GetName()), 0777)
		fhandle, err := os.Create(filename)
		if err != nil {
			s.Log(fmt.Sprintf("Unable to open file: %v", err))
		}
		s.dlogHandle = fhandle
	} else {
		s.Log(fmt.Sprintf("Scratch not found, no disk logging"))
	}
}
