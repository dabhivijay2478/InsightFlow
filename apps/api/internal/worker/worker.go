package worker

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/mantrixflow/go-api/internal/config"
	"github.com/mantrixflow/go-api/internal/queue"
	"github.com/mantrixflow/go-api/internal/server"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	cfg *config.Config
	s   *server.State
	q   *queue.Service
}

func New(cfg *config.Config, st *server.State, q *queue.Service) *Worker {
	return &Worker{cfg: cfg, s: st, q: q}
}

func (w *Worker) Run(ctx context.Context) {
	if w.q == nil {
		log.Warn().Msg("pgmq disabled — worker not started")
		return
	}
	tick := time.NewTicker(time.Duration(w.cfg.PGMQPollIntervalMS) * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			w.poll(ctx, queue.QueuePipelineJobs, w.cfg.PGMQVTLongSec, w.cfg.PGMQParallelWorkers)
			w.poll(ctx, queue.QueueIncrementalSync, w.cfg.PGMQVTLongSec, w.cfg.PGMQParallelWorkers)
			w.poll(ctx, queue.QueuePollingChecks, w.cfg.PGMQVTShortSec, 1)
		}
	}
}

func (w *Worker) poll(ctx context.Context, qname string, vt int, batch int) {
	msgs, err := w.q.Read(ctx, qname, vt, batch)
	if err != nil {
		log.Warn().Err(err).Str("queue", qname).Msg("pgmq read")
		return
	}
	for _, msg := range msgs {
		m := msg
		qn := qname
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error().Interface("panic", r).Bytes("stack", debug.Stack()).Str("queue", qn).Msg("dispatch panic")
				}
			}()
			c, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()
			var err error
			switch qn {
			case queue.QueuePipelineJobs:
				err = w.s.DispatchFullSyncJob(c, w.q, m)
			case queue.QueueIncrementalSync:
				err = w.s.DispatchIncrementalSyncJob(c, w.q, m)
			case queue.QueuePollingChecks:
				err = w.s.DispatchPollingJob(c, w.q, m)
			}
			if err != nil {
				log.Warn().Err(err).Str("queue", qn).Msg("dispatch")
			}
		}()
	}
}
