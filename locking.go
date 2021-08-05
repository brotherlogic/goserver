package goserver

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	lpb "github.com/brotherlogic/lock/proto"
)

var election = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "server_election",
	Help: "The number of active election waits",
})

//Elect elect me
func (s *GoServer) Elect() (func(), error) {
	if s.SkipElect {
		return func() {

		}, nil
	}
	elected := make(chan error, 1)
	complete := make(chan bool, 1)
	rf := func() {
		complete <- true
	}

	go s.runElection("", elected, complete)

	err := <-elected
	return rf, err
}

//ElectKey elect me with a key
func (s *GoServer) ElectKey(key string) (func(), error) {
	if s.SkipElect {
		return func() {

		}, nil
	}
	elected := make(chan error)
	complete := make(chan bool)
	rf := func() {
		complete <- true
	}

	go s.runElection(key, elected, complete)

	err := <-elected
	return rf, err
}

func (s *GoServer) runElection(key string, elected chan error, complete chan bool) {
	s.RaiseIssue(fmt.Sprintf("%v is running election", s.Registry.Identifier), fmt.Sprintf("The key is %v", key))
	if s.SkipElect {
		elected <- nil
		return
	}
	election.Inc()
	defer election.Dec()
	command := exec.Command("etcdctl", "elect", s.Registry.Name+key, s.Registry.Identifier)
	command.Env = append(os.Environ(),
		"ETCDCTL_API=3",
	)
	out, _ := command.StdoutPipe()
	defer out.Close()
	if out != nil {
		scanner := bufio.NewScanner(out)
		go func() {
			for scanner != nil && scanner.Scan() {
				text := scanner.Text()
				// Expect something like registry.name/key
				if strings.HasPrefix(text, s.Registry.Name) {
					elected <- nil
				} else {
					elected <- fmt.Errorf("Unable to elect (%v): %v", s.Registry.Name+key, scanner.Text())
				}
			}
		}()
	}

	out2, _ := command.StderrPipe()
	defer out2.Close()
	if out2 != nil {
		scanner := bufio.NewScanner(out2)
		go func() {
			for scanner != nil && scanner.Scan() {
				text := scanner.Text()
				s.Log(fmt.Sprintf("ERR run %v", text))
				if strings.HasPrefix(text, s.Registry.Name) {
					elected <- nil
				} else {
					elected <- fmt.Errorf("Unable to from err elect (%v): %v", s.Registry.Name+key, scanner.Text())
				}
			}
		}()
	}

	//Run the election
	err := command.Start()
	if err != nil {
		s.Log(fmt.Sprintf("Error starting command: %v", err))
	}
	<-complete
	command.Process.Kill()

	// Ensure that the command is stopped an removed
	command.Wait()
}

func (s *GoServer) RunLockingElection(ctx context.Context, key string) (string, error) {
	conn, err := s.FDialServer(ctx, "lock")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	client := lpb.NewLockServiceClient(conn)

	// Default to a 60 second deadline
	duration := time.Second * 60
	dead, ok := ctx.Deadline()
	if ok {
		duration = time.Until(dead)
	}

	res, err := client.AcquireLock(ctx, &lpb.AcquireLockRequest{
		Key:                   key,
		LockDurationInSeconds: int64(duration.Seconds()),
	})
	if err != nil {
		return "", err
	}

	return res.GetLock().GetLockKey(), nil
}

func (s *GoServer) ReleaseLockingElection(ctx context.Context, key string, lockKey string) error {
	conn, err := s.FDialServer(ctx, "lock")
	if err != nil {
		return err
	}
	defer conn.Close()
	client := lpb.NewLockServiceClient(conn)

	_, err = client.ReleaseLock(ctx, &lpb.ReleaseLockRequest{
		Key:     key,
		LockKey: lockKey,
	})
	return err
}
