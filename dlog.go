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
	file, _ := os.Open("/proc/mounts")
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
	badWrites = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "dlog_bad_write",
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

// DLog writes to the dlog
func (s *GoServer) DLog(ctx context.Context, text string) {
	if s.dlogHandle != nil {
		code, err := utils.GetContextKey(ctx)
		if err != nil {
			server := "unknown"
			if s.Registry != nil {
				server = s.Registry.Identifier
			}
			badWrites.Inc()
			s.RaiseIssue("Logging error", fmt.Sprintf("Log line %v had no context key (%v) -> %v", text, server, err))
			code = "NONE"
		}
		s.dlogHandle.WriteString(fmt.Sprintf("%v|%v|%v|%v\n", time.Now().Format(time.RFC3339Nano), s.Registry.GetIdentifier(), code, text))

		if time.Since(s.lastLogCheck) > time.Minute {
			if _, err := os.Stat(fmt.Sprintf("/media/scratch/dlogs/%v", s.Registry.GetName())); !os.IsNotExist(err) {
				size, err := dirSize(fmt.Sprintf("/media/scratch/dlogs/%v", s.Registry.GetName()))
				if err != nil {
					if s.Registry != nil {
						s.RaiseIssue("Bad log problem", fmt.Sprintf("Error reeading logs on %v: %v", s.Registry.Identifier, err))
					}
				} else {
					s.lastLogCheck = time.Now()
					logSize.Set(float64(size))
				}
			}
		}
	}
}

func (s *GoServer) prepDLog(serviceName string) {
	s.preppedDLog = true
	if s.hasScratch() {
		filename := fmt.Sprintf("/media/scratch/dlogs/%v/%v.logs", serviceName, time.Now().Unix())
		os.MkdirAll(fmt.Sprintf("/media/scratch/dlogs/%v/", serviceName), 0777)

		//Delete all files a week old
		files, _ := ioutil.ReadDir(fmt.Sprintf("/media/scratch/dlogs/%v/", serviceName))

		count := 0
		for _, file := range files {
			if time.Since(file.ModTime()) > time.Hour*24*7 {
				os.Remove(fmt.Sprintf("/media/scratch/dlogs/%v/%v", serviceName, file.Name()))
				count++
			}
		}

		fhandle, _ := os.Create(filename)
		s.dlogHandle = fhandle
	} else if serviceName == "gobuildslave" {
		filename := fmt.Sprintf("/home/simon/gobuildslave.tlog")
		fhandle, _ := os.Create(filename)
		s.dlogHandle = fhandle
	} else {
		hn, err := os.Hostname()
		s.RaiseIssue("Missing Disk Logs", fmt.Sprintf("%v,%v,%v has not disk logging potential", hn, err, s.Registry.GetIdentifier()))
	}
}
