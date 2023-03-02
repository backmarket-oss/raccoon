package internal

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type Strategy interface {
	//Run is a blocking function
	Run(ctx context.Context) error
}

type DefaultSettings struct {
	Namespace string
	Selector  string
	TTL       time.Duration
	DryRun    bool
}

// RunDaemon is the main loop driven by a check interval.
func RunDaemon(interval int, ctx context.Context, stg Strategy) error {
	for {
		log.Debug("Racoon, wake up")
		err := stg.Run(ctx)
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			log.Debug("Raccoon, stop")
			return nil
		case <-time.After(time.Duration(interval) * time.Second):
		}
	}
}
