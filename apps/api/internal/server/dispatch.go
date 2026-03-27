package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mantrixflow/go-api/internal/queue"
	"github.com/rs/zerolog/log"
)

func (s *State) loadConnection(orgID, dataSourceID string) (connType string, cfg map[string]interface{}, status string, replName *string, err error) {
	var cfgRaw []byte
	var slot sql.NullString
	row := s.DB.Raw(`
		SELECT c.connection_type, c.config, c.status, c.replication_slot_name
		FROM data_source_connections c
		INNER JOIN data_sources ds ON ds.id = c.data_source_id
		WHERE c.data_source_id = ?::uuid AND ds.organization_id = ?::uuid
		LIMIT 1`,
		dataSourceID, orgID).Row()
	if row.Err() != nil {
		return "", nil, "", nil, row.Err()
	}
	if err := row.Scan(&connType, &cfgRaw, &status, &slot); err != nil {
		return "", nil, "", nil, err
	}
	if err := json.Unmarshal(cfgRaw, &cfg); err != nil {
		return "", nil, "", nil, err
	}
	if slot.Valid && slot.String != "" {
		replName = &slot.String
	}
	return connType, cfg, status, replName, nil
}

func (s *State) decryptedConnMap(orgID, dataSourceID, connType string) (map[string]interface{}, error) {
	_, cfg, _, _, err := s.loadConnection(orgID, dataSourceID)
	if err != nil {
		return nil, err
	}
	return s.DecryptConnectionJSON(connType, cfg)
}

// DispatchFullSyncJob processes a PGMQ full-sync message (same contract as Nest).
func (s *State) DispatchFullSyncJob(ctx context.Context, q *queue.Service, msg queue.Message) error {
	var wrapper queue.JobPayload
	if err := json.Unmarshal(msg.Message, &wrapper); err != nil {
		return err
	}
	var job queue.FullSyncData
	if err := json.Unmarshal(wrapper.Data, &job); err != nil {
		return err
	}
	pipelineID := job.PipelineID
	runID := job.RunID
	orgID := job.OrganizationID
	userID := job.UserID
	if userID == "" {
		userID = "system"
	}

	pipe, src, dest, err := s.loadPipelineBundle(nil, pipelineID)
	if err != nil {
		return err
	}
	destEff := effectiveDestinationSchema(pipe, dest)

	srcDS := strField(src, "data_source_id")
	dstDS := strField(destEff, "data_source_id")
	if srcDS == "" || dstDS == "" {
		return fmt.Errorf("missing data source on schema")
	}

	srcTable := strField(src, "source_table")
	if srcTable == "" || strings.EqualFold(srcTable, "placeholder") {
		_ = s.failRun(runID, "Pipeline source table is not configured. Select a source table in the pipeline builder before running.")
		_ = q.Archive(ctx, queue.QueuePipelineJobs, msg.MsgID)
		return nil
	}

	sConn := s.loadConnRow(orgID, srcDS)
	dConn := s.loadConnRow(orgID, dstDS)
	if sConn == nil || dConn == nil {
		msgErr := "Source or destination has no connection. Configure and connect the data sources first."
		_ = s.failRun(runID, msgErr)
		_ = q.Archive(ctx, queue.QueuePipelineJobs, msg.MsgID)
		return nil
	}
	if sConn.status == "error" || dConn.status == "error" {
		msgErr := "Source or destination connection has errors. Fix the connection and try again."
		_ = s.failRun(runID, msgErr)
		_ = q.Archive(ctx, queue.QueuePipelineJobs, msg.MsgID)
		return nil
	}

	srcCfg, err := s.DecryptConnectionJSON(sConn.connType, sConn.cfg)
	if err != nil {
		return err
	}
	destCfg, err := s.DecryptConnectionJSON(dConn.connType, dConn.cfg)
	if err != nil {
		return err
	}

	syncMode := strField(pipe, "sync_mode")
	if syncMode == "" {
		syncMode = "full"
	}
	effectiveSync := syncMode
	if syncMode == "incremental" {
		effectiveSync = "incremental"
	} else {
		effectiveSync = "full"
	}
	repKey := strField(pipe, "incremental_column")
	if effectiveSync != "incremental" {
		repKey = ""
	}

	_ = s.updateRun(runID, map[string]interface{}{
		"status":     "running",
		"job_state":  "running",
		"started_at": time.Now(),
		"updated_at": time.Now(),
	})
	_ = s.updatePipeline(pipelineID, map[string]interface{}{
		"status":      "running",
		"last_run_at": time.Now(),
		"updated_at":  time.Now(),
	})

	colMap := extractColumnMap(pipe)
	txScript := extractTransformScript(pipe, destEff)
	writeMode := strField(destEff, "write_mode")
	if writeMode == "" {
		writeMode = "upsert"
	}
	emit := "upsert"
	if writeMode == "replace" {
		emit = "replace"
	} else if writeMode == "append" {
		emit = "append"
	}

	srcSchemaName := strField(src, "source_schema")
	if srcSchemaName == "" {
		srcSchemaName = "public"
	}
	destSchemaName := strField(destEff, "destination_schema")
	destTableName := strField(destEff, "destination_table")
	sourceStream := srcTable
	if srcSchemaName != "" && srcTable != "" {
		sourceStream = srcSchemaName + "-" + srcTable
	}

	payload := map[string]interface{}{
		"job_id":                   runID,
		"pipeline_id":              pipelineID,
		"organization_id":          orgID,
		"source_connection_config": srcCfg,
		"dest_connection_config":   destCfg,
		"source_type":              registrySourceType(sConn.connType),
		"dest_type":                registrySourceType(dConn.connType),
		"replication_method":       replicationMethodFromSyncMode(effectiveSync),
		"source_stream":            sourceStream,
		"dest_table":               destTableName,
		"dest_schema":              destSchemaName,
		"column_map":               colMap,
		"transform_script":         nil,
		"emit_method":              emit,
		"upsert_key":               destEff["upsert_key"],
		"hard_delete":              false,
		"replication_key":          nil,
		"replication_slot_name":    nil,
	}
	if repKey != "" {
		payload["replication_key"] = repKey
	}
	if txScript != "" {
		payload["transform_script"] = txScript
	}

	retry, err := s.ETL.RunSync(s.Cfg.APIPublicURL, payload)
	if err != nil {
		_ = s.failRun(runID, err.Error())
		_ = s.handleDispatchFailure(ctx, q, queue.QueuePipelineJobs, msg, wrapper)
		return nil
	}
	if retry {
		_ = s.updateRun(runID, map[string]interface{}{"status": "pending", "job_state": "queued", "updated_at": time.Now()})
		_ = s.updatePipeline(pipelineID, map[string]interface{}{"status": "idle", "updated_at": time.Now()})
		_ = q.Archive(ctx, queue.QueuePipelineJobs, msg.MsgID)
		if wrapper.RetryCount < wrapper.MaxRetries {
			_ = queue.RequeueWithBackoff(ctx, q, queue.QueuePipelineJobs, wrapper)
		} else {
			_ = s.failRun(runID, "Exhausted dispatch retries — all ETL pods at capacity")
		}
		return nil
	}
	log.Info().Str("pipeline", pipelineID).Str("run", runID).Msg("full sync dispatched to ETL")
	_ = q.Archive(ctx, queue.QueuePipelineJobs, msg.MsgID)
	q.PublishStatus(ctx, pipelineID, orgID, "running", intPtr(0), nil)
	return nil
}

type connRow struct {
	connType string
	cfg      map[string]interface{}
	status   string
	repl     *string
}

func (s *State) loadConnRow(orgID, ds string) *connRow {
	ct, cfg, st, rp, err := s.loadConnection(orgID, ds)
	if err != nil {
		return nil
	}
	return &connRow{connType: ct, cfg: cfg, status: st, repl: rp}
}

func intPtr(i int) *int { return &i }

func (s *State) updateRun(runID string, fields map[string]interface{}) error {
	if runID == "" {
		return nil
	}
	return s.DB.Table("pipeline_runs").Where("id = ?::uuid", runID).Updates(fields).Error
}

func (s *State) updatePipeline(id string, fields map[string]interface{}) error {
	return s.DB.Table("pipelines").Where("id = ?::uuid", id).Updates(fields).Error
}

func (s *State) failRun(runID, msg string) error {
	if runID == "" {
		return nil
	}
	return s.updateRun(runID, map[string]interface{}{
		"status":        "failed",
		"job_state":     "failed",
		"error_message": msg,
		"completed_at":  time.Now(),
		"updated_at":    time.Now(),
	})
}

func (s *State) handleDispatchFailure(ctx context.Context, q *queue.Service, qname string, msg queue.Message, wrapper queue.JobPayload) error {
	_ = q.Archive(ctx, qname, msg.MsgID)
	if wrapper.RetryCount < wrapper.MaxRetries {
		return queue.RequeueWithBackoff(ctx, q, qname, wrapper)
	}
	return nil
}
