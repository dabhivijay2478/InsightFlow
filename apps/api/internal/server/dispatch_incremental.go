package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mantrixflow/go-api/internal/queue"
	"github.com/rs/zerolog/log"
)

func (s *State) DispatchIncrementalSyncJob(ctx context.Context, q *queue.Service, msg queue.Message) error {
	var wrapper queue.JobPayload
	if err := json.Unmarshal(msg.Message, &wrapper); err != nil {
		return err
	}
	var job queue.IncrementalSyncData
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
	st := strField(pipe, "status")
	if st != "listing" && st != "idle" && st != "completed" && st != "failed" {
		_ = q.Archive(ctx, queue.QueueIncrementalSync, msg.MsgID)
		return nil
	}

	sm := strField(pipe, "sync_mode")
	isCDC := sm == "cdc" || sm == "log_based"
	if isCDC && !s.Cfg.AllowSourceDBMutationsForCDC {
		_ = s.failRun(runID, "CDC and LOG_BASED sync are disabled by policy because they can alter the client source database (for example, replication slots).")
		_ = q.Archive(ctx, queue.QueueIncrementalSync, msg.MsgID)
		return nil
	}
	if isCDC {
		if strField(pipe, "full_refresh_completed_at") == "" {
			_ = s.failRun(runID, "This pipeline requires an initial full sync before log-based sync can run. Click 'Run Initial Sync' to start.")
			_ = q.Archive(ctx, queue.QueueIncrementalSync, msg.MsgID)
			return nil
		}
	}

	destEff := effectiveDestinationSchema(pipe, dest)
	srcDS := strField(src, "data_source_id")
	dstDS := strField(destEff, "data_source_id")
	srcTable := strField(src, "source_table")
	if srcTable == "" || strings.EqualFold(srcTable, "placeholder") {
		_ = s.failRun(runID, "Pipeline source table is not configured.")
		_ = q.Archive(ctx, queue.QueueIncrementalSync, msg.MsgID)
		return nil
	}

	sConn := s.loadConnRow(orgID, srcDS)
	dConn := s.loadConnRow(orgID, dstDS)
	if sConn == nil || dConn == nil || sConn.status == "error" || dConn.status == "error" {
		_ = s.failRun(runID, "Source or destination connection missing or in error state.")
		_ = q.Archive(ctx, queue.QueueIncrementalSync, msg.MsgID)
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

	repl := ""
	if sConn.repl != nil && *sConn.repl != "" {
		repl = *sConn.repl
	} else if sConn != nil {
		// derive like Nest: mxf_ + first 8 hex of connection id — we use data source id slice
		repl = "mxf_" + strings.ReplaceAll(srcDS, "-", "")[:8]
	}

	_ = s.updateRun(runID, map[string]interface{}{
		"status": "running", "job_state": "running", "started_at": time.Now(), "updated_at": time.Now(),
	})
	_ = s.updatePipeline(pipelineID, map[string]interface{}{
		"status": "running", "last_run_at": time.Now(), "updated_at": time.Now(),
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
	sourceStream := srcSchemaName + "-" + srcTable

	etlSyncMode := "cdc"
	if strField(pipe, "sync_mode") == "incremental" {
		etlSyncMode = "incremental"
	}

	payload := map[string]interface{}{
		"job_id":                   runID,
		"pipeline_id":              pipelineID,
		"organization_id":          orgID,
		"source_connection_config": srcCfg,
		"dest_connection_config":   destCfg,
		"source_type":              registrySourceType(sConn.connType),
		"dest_type":                registrySourceType(dConn.connType),
		"replication_method":       replicationMethodFromSyncMode(etlSyncMode),
		"source_stream":            sourceStream,
		"dest_table":               strField(destEff, "destination_table"),
		"dest_schema":              strField(destEff, "destination_schema"),
		"column_map":               colMap,
		"transform_script":         nil,
		"emit_method":              emit,
		"upsert_key":               destEff["upsert_key"],
		"hard_delete":              false,
		"replication_slot_name":    repl,
	}
	if txScript != "" {
		payload["transform_script"] = txScript
	}

	retry, err := s.ETL.RunSync(s.Cfg.APIPublicURL, payload)
	if err != nil {
		_ = s.failRun(runID, err.Error())
		_ = s.handleDispatchFailure(ctx, q, queue.QueueIncrementalSync, msg, wrapper)
		return nil
	}
	if retry {
		_ = s.updateRun(runID, map[string]interface{}{"status": "pending", "job_state": "queued", "updated_at": time.Now()})
		_ = s.updatePipeline(pipelineID, map[string]interface{}{"status": "listing", "updated_at": time.Now()})
		_ = q.Archive(ctx, queue.QueueIncrementalSync, msg.MsgID)
		if wrapper.RetryCount < wrapper.MaxRetries {
			_ = queue.RequeueWithBackoff(ctx, q, queue.QueueIncrementalSync, wrapper)
		} else {
			_ = s.failRun(runID, "Exhausted dispatch retries — all ETL pods at capacity")
		}
		return nil
	}
	log.Info().Str("pipeline", pipelineID).Str("run", runID).Msg("incremental sync dispatched")
	_ = q.Archive(ctx, queue.QueueIncrementalSync, msg.MsgID)
	return nil
}

func (s *State) DispatchPollingJob(ctx context.Context, q *queue.Service, msg queue.Message) error {
	var wrapper queue.JobPayload
	if err := json.Unmarshal(msg.Message, &wrapper); err != nil {
		return err
	}
	if wrapper.Name == "poll-cycle" {
		var rows []struct {
			ID             string  `gorm:"column:id"`
			OrganizationID string  `gorm:"column:organization_id"`
			CreatedBy      *string `gorm:"column:created_by"`
		}
		_ = s.DB.Raw(`
			SELECT id, organization_id, created_by FROM pipelines
			WHERE deleted_at IS NULL AND status = 'listing'`).Scan(&rows).Error
		for _, r := range rows {
			_ = queue.EnqueueDeltaCheck(ctx, q, queue.DeltaCheckData{
				PipelineID:     r.ID,
				OrganizationID: r.OrganizationID,
			})
		}
		_ = q.Archive(ctx, queue.QueuePollingChecks, msg.MsgID)
		return nil
	}
	var d queue.DeltaCheckData
	if err := json.Unmarshal(wrapper.Data, &d); err != nil {
		_ = q.Archive(ctx, queue.QueuePollingChecks, msg.MsgID)
		return err
	}
	pipe, _, _, err := s.loadPipelineBundle(nil, d.PipelineID)
	if err != nil || strField(pipe, "status") != "listing" {
		_ = q.Archive(ctx, queue.QueuePollingChecks, msg.MsgID)
		return nil
	}
	sm := strField(pipe, "sync_mode")
	if sm != "cdc" && sm != "log_based" {
		_ = q.Archive(ctx, queue.QueuePollingChecks, msg.MsgID)
		return nil
	}
	uid := "system"
	if cb := strField(pipe, "created_by"); cb != "" {
		uid = cb
	}
	runID := ""
	// create run
	type runRow struct {
		ID string `gorm:"column:id"`
	}
	var rr runRow
	var insErr error
	if uid == "" || uid == "system" {
		insErr = s.DB.Raw(`
			INSERT INTO pipeline_runs (pipeline_id, organization_id, status, job_state, trigger_type, triggered_by, started_at, created_at, updated_at)
			VALUES (?::uuid, ?::uuid, 'pending', 'queued', 'polling', NULL, NOW(), NOW(), NOW()) RETURNING id`,
			d.PipelineID, d.OrganizationID).Scan(&rr).Error
	} else {
		insErr = s.DB.Raw(`
			INSERT INTO pipeline_runs (pipeline_id, organization_id, status, job_state, trigger_type, triggered_by, started_at, created_at, updated_at)
			VALUES (?::uuid, ?::uuid, 'pending', 'queued', 'polling', ?::uuid, NOW(), NOW(), NOW()) RETURNING id`,
			d.PipelineID, d.OrganizationID, uid).Scan(&rr).Error
	}
	if insErr == nil {
		runID = rr.ID
	}
	if runID == "" {
		_ = q.Archive(ctx, queue.QueuePollingChecks, msg.MsgID)
		return fmt.Errorf("create run")
	}
	_ = queue.EnqueueIncrementalSync(ctx, q, queue.IncrementalSyncData{
		PipelineID:     d.PipelineID,
		RunID:          runID,
		OrganizationID: d.OrganizationID,
		UserID:         uid,
		TriggerType:    "polling",
		BatchSize:      500,
	})
	_ = q.Archive(ctx, queue.QueuePollingChecks, msg.MsgID)
	return nil
}
