package goserver

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	lpb "github.com/brotherlogic/lock/proto"
)

var election = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "server_election",
	Help: "The number of active election waits",
})

func (s *GoServer) RunLockingElection(ctx context.Context, key string, detail string) (string, error) {
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
		Purpose:               detail,
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
