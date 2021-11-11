package goserver

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brotherlogic/goserver/utils"
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

//DLog writes to the dlog
func (s *GoServer) DLog(ctx context.Context, text string) {
	if s.dlogHandle != nil {
		code, err := utils.GetContextKey(ctx)
		if err != nil {
			code = "NONE"
		}
		s.dlogHandle.WriteString(fmt.Sprintf("%v|%v|%v|%v\n", time.Now().Format(time.RFC3339Nano), s.Registry.GetIdentifier(), code, text))

		size, err := dirSize(fmt.Sprintf("/media/scratch/dlogs/%v", s.Registry.GetName()))
		if err != nil {
			s.RaiseIssue("Bad log problem", fmt.Sprintf("Error reeading logs: %v", err))
		} else {
			logSize.Set(float64(size))
		}

	}
}

func (s *GoServer) prepDLog(serviceName string) {
	s.preppedDLog = true
	if s.hasScratch() {
		filename := fmt.Sprintf("/media/scratch/dlogs/%v/%v.logs", serviceName, time.Now().Unix())
		err := os.MkdirAll(fmt.Sprintf("/media/scratch/dlogs/%v/", serviceName), 0777)
		if err != nil {
			s.Log(fmt.Sprintf("Unable to create log dir: %v", err))
		}

		//Delete all files a week old
		files, err := ioutil.ReadDir(fmt.Sprintf("/media/scratch/dlogs/%v/", serviceName))
		if err != nil {
			s.Log(fmt.Sprintf("Unable to list log files: %v", err))
		}
		count := 0
		for _, file := range files {
			if time.Since(file.ModTime()) > time.Hour*24*7 {
				os.Remove(fmt.Sprintf("/media/scratch/dlogs/%v/%v", serviceName, file.Name()))
				count++
			}
		}
		s.Log(fmt.Sprintf("Removed %v files", count))

		fhandle, err := os.Create(filename)
		if err != nil {
			s.Log(fmt.Sprintf("Unable to open file: %v", err))
		}
		s.dlogHandle = fhandle
		s.Log(fmt.Sprintf("Prepped dlog"))
	} else {
		s.Log(fmt.Sprintf("Scratch not found, no disk logging"))
		hn, err := os.Hostname()
		s.RaiseIssue("Missing Disk Logs", fmt.Sprintf("%v,%v,%v has not disk logging potential", hn, err, s.Registry.GetIdentifier()))
	}
}
