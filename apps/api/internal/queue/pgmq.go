package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mantrixflow/go-api/internal/config"
	"github.com/rs/zerolog/log"
)

const (
	QueuePipelineJobs    = "pipeline_jobs"
	QueueIncrementalSync = "incremental_sync"
	QueuePollingChecks   = "polling_checks"
	NotifyPipelineStatus = "pipeline_job_status"
	CronCdcPollJob       = "pgmq_cdc_poll_cycle"
)

type JobPayload struct {
	Name       string          `json:"name"`
	Data       json.RawMessage `json:"data"`
	RetryCount int             `json:"retryCount"`
	MaxRetries int             `json:"maxRetries"`
}

type Message struct {
	MsgID      int64           `json:"msg_id"`
	ReadCt     int32           `json:"read_ct"`
	EnqueuedAt interface{}     `json:"enqueued_at"`
	VT         interface{}     `json:"vt"`
	Message    json.RawMessage `json:"message"`
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(ctx context.Context, cfg *config.Config) (*Service, error) {
	url := cfg.SessionDatabaseURL()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("pgx pool: %w", err)
	}
	s := &Service{pool: pool}
	if err := s.bootstrap(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *Service) Close() { s.pool.Close() }

func (s *Service) bootstrap(ctx context.Context) error {
	_, _ = s.pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgmq`)
	for _, name := range []string{QueuePipelineJobs, QueueIncrementalSync, QueuePollingChecks} {
		_, err := s.pool.Exec(ctx, `SELECT pgmq.create($1)`, name)
		if err != nil {
			msg := strings.ToLower(err.Error())
			if strings.Contains(msg, "exists") || strings.Contains(msg, "duplicate") {
				continue
			}
			log.Warn().Err(err).Str("queue", name).Msg("pgmq create queue")
		}
	}
	// pg_cron optional — same as Nest
	_, _ = s.pool.Exec(ctx, `SELECT cron.unschedule($1)`, CronCdcPollJob)
	_, err := s.pool.Exec(ctx, `SELECT cron.schedule($1, $2, $3)`,
		CronCdcPollJob, `*/5 * * * *`,
		`SELECT pgmq.send('polling_checks', '{"name":"poll-cycle","data":{},"retryCount":0,"maxRetries":1}'::jsonb)`)
	if err != nil {
		log.Warn().Err(err).Msg("pg_cron schedule optional failed")
	}
	return nil
}

func (s *Service) Read(ctx context.Context, queue string, vtSec int, qty int) ([]Message, error) {
	rows, err := s.pool.Query(ctx, `SELECT msg_id, read_ct, enqueued_at, vt, message FROM pgmq.read($1, $2::integer, $3::integer)`, queue, vtSec, qty)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.MsgID, &m.ReadCt, &m.EnqueuedAt, &m.VT, &m.Message); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Service) Archive(ctx context.Context, queue string, msgID int64) error {
	_, err := s.pool.Exec(ctx, `SELECT pgmq.archive($1, $2::bigint)`, queue, msgID)
	return err
}

func (s *Service) Delete(ctx context.Context, queue string, msgID int64) error {
	_, err := s.pool.Exec(ctx, `SELECT pgmq.delete($1, $2::bigint)`, queue, msgID)
	return err
}

func (s *Service) Send(ctx context.Context, queue string, payload JobPayload) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `SELECT * FROM pgmq.send($1, $2::jsonb)`, queue, string(b))
	return err
}

func (s *Service) SendDelay(ctx context.Context, queue string, payload JobPayload, delaySec int) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `SELECT * FROM pgmq.send($1, $2::jsonb, $3::integer)`, queue, string(b), delaySec)
	return err
}

func (s *Service) PublishStatus(ctx context.Context, pipelineID, organizationID, status string, rows *int, errMsg *string) {
	ev := map[string]interface{}{
		"pipelineId":     pipelineID,
		"organizationId": organizationID,
		"status":         status,
		"timestamp":      "",
	}
	if rows != nil {
		ev["rowsProcessed"] = *rows
	}
	if errMsg != nil {
		ev["error"] = *errMsg
	}
	b, _ := json.Marshal(ev)
	payload := string(b)
	if len(payload) > 7900 {
		payload = payload[:7900]
	}
	_, _ = s.pool.Exec(ctx, `SELECT pg_notify($1, $2)`, NotifyPipelineStatus, payload)
}

type FullSyncData struct {
	PipelineID     string `json:"pipelineId"`
	RunID          string `json:"runId"`
	OrganizationID string `json:"organizationId"`
	UserID         string `json:"userId"`
	TriggerType    string `json:"triggerType"`
	BatchSize      int    `json:"batchSize,omitempty"`
}

type IncrementalSyncData struct {
	PipelineID     string                 `json:"pipelineId"`
	RunID          string                 `json:"runId"`
	OrganizationID string                 `json:"organizationId"`
	UserID         string                 `json:"userId"`
	TriggerType    string                 `json:"triggerType"`
	Checkpoint     map[string]interface{} `json:"checkpoint,omitempty"`
	BatchSize      int                    `json:"batchSize,omitempty"`
}

type DeltaCheckData struct {
	PipelineID     string `json:"pipelineId"`
	OrganizationID string `json:"organizationId"`
}

func EnqueueFullSync(ctx context.Context, q *Service, d FullSyncData) error {
	p := JobPayload{Name: "full-sync", RetryCount: 0, MaxRetries: 5}
	db, _ := json.Marshal(d)
	p.Data = db
	return q.Send(ctx, QueuePipelineJobs, p)
}

func EnqueueIncrementalSync(ctx context.Context, q *Service, d IncrementalSyncData) error {
	p := JobPayload{Name: "incremental-sync", RetryCount: 0, MaxRetries: 5}
	db, _ := json.Marshal(d)
	p.Data = db
	return q.Send(ctx, QueueIncrementalSync, p)
}

// RequeueWithBackoff mirrors Nest pgmq exponential delay.
func RequeueWithBackoff(ctx context.Context, q *Service, queueName string, payload JobPayload) error {
	delay := 1 << minInt(payload.RetryCount+1, 8)
	if delay > 300 {
		delay = 300
	}
	p := payload
	p.RetryCount++
	return q.SendDelay(ctx, queueName, p, delay)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func EnqueueDeltaCheck(ctx context.Context, q *Service, d DeltaCheckData) error {
	p := JobPayload{Name: "delta-check", RetryCount: 0, MaxRetries: 1}
	db, _ := json.Marshal(d)
	p.Data = db
	return q.Send(ctx, QueuePollingChecks, p)
}
