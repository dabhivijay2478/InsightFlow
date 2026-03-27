package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mantrixflow/go-api/pkg/response"
)

func sanitizeETLError(msg string) string {
	msg = strings.TrimSpace(msg)
	if len(msg) > 2000 {
		return msg[:2000]
	}
	return msg
}

func checkpointLastSyncValue(cp map[string]interface{}) interface{} {
	for _, k := range []string{"lastSyncValue", "last_value", "last_commit_lsn", "lsnEnd", "lsn"} {
		if v, ok := cp[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func checkpointLastSyncAt(cp map[string]interface{}) string {
	for _, k := range []string{"lastSyncAt", "updated_at", "synced_at"} {
		if v, ok := cp[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// EtlCallback handles POST /internal/etl-callback (InternalEtlController parity, email omitted).
func (s *State) EtlCallback(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
	}
	runID := strField(body, "run_id")
	if runID == "" {
		runID = strField(body, "job_id")
	}
	rowsRead := num(body["rows_read"])
	rowsUpserted := num(body["rows_upserted"])
	if rowsUpserted == 0 {
		rowsUpserted = num(body["rows_written"])
	}
	rowsDropped := num(body["rows_dropped"])
	rowsDeleted := num(body["rows_deleted"])
	errMsg := strField(body, "error_message")
	if errMsg == "" {
		errMsg = strField(body, "error")
	}
	dur := num(body["duration_seconds"])
	srcTool := strField(body, "source_tool")
	dstTool := strField(body, "dest_tool")
	repMethod := strField(body, "replication_method_used")
	rawStatus := strings.ToLower(strField(body, "status"))
	if rawStatus == "" {
		rawStatus = "failed"
	}
	status := "failed"
	if rawStatus == "completed" || rawStatus == "success" || rawStatus == "partial_success" {
		status = "completed"
	} else if rawStatus == "interrupted" {
		status = "interrupted"
	}

	cp, _ := body["checkpoint"].(map[string]interface{})
	if cp == nil && body["last_cursor_value"] != nil {
		if m, ok := body["last_cursor_value"].(map[string]interface{}); ok {
			cp = m
		}
	}

	pipelineID := strField(body, "pipeline_id")
	if pipelineID == "" && runID != "" {
		var pid string
		_ = s.DB.Raw(`SELECT pipeline_id::text FROM pipeline_runs WHERE id = ?::uuid LIMIT 1`, runID).Scan(&pid)
		pipelineID = pid
	}
	if pipelineID == "" {
		return c.Status(http.StatusOK).JSON(fiber.Map{"received": false, "error": "pipeline_id missing"})
	}

	var pj []byte
	if err := s.DB.Raw(`SELECT to_jsonb(p) FROM pipelines p WHERE p.id = ?::uuid LIMIT 1`, pipelineID).Scan(&pj).Error; err != nil || len(pj) == 0 {
		return c.Status(http.StatusOK).JSON(fiber.Map{"received": true, "warning": "Pipeline not found"})
	}
	var pipe map[string]interface{}
	if err := json.Unmarshal(pj, &pipe); err != nil {
		return c.Status(http.StatusOK).JSON(fiber.Map{"received": true, "warning": "Pipeline not found"})
	}

	runStatus := "failed"
	jobState := "completed"
	if status == "completed" {
		runStatus = "success"
	} else if status == "interrupted" {
		jobState = "failed"
	} else {
		jobState = "completed"
	}
	san := sanitizeETLError(errMsg)
	rowsFailed := 0
	if status != "completed" {
		rowsFailed = int(math.Max(float64(rowsRead-rowsUpserted-rowsDropped), 1))
	}

	_ = s.DB.Table("pipeline_runs").Where("id = ?::uuid", runID).Updates(map[string]interface{}{
		"status":                 runStatus,
		"job_state":              jobState,
		"rows_read":              rowsRead,
		"rows_written":           rowsUpserted,
		"rows_skipped":           rowsDropped,
		"rows_failed":            rowsFailed,
		"rows_deleted":           rowsDeleted,
		"source_tool":            srcTool,
		"dest_tool":              dstTool,
		"collection_method_used": repMethod,
		"completed_at":           time.Now(),
		"duration_seconds":       dur,
		"error_message":          san,
		"updated_at":             time.Now(),
	}).Error

	if cp != nil && status == "completed" {
		lsv := checkpointLastSyncValue(cp)
		lsa := checkpointLastSyncAt(cp)
		up := map[string]interface{}{
			"checkpoint":      cp,
			"last_sync_value": nil,
			"updated_at":      time.Now(),
		}
		if lsv != nil {
			up["last_sync_value"] = fmt.Sprint(lsv)
		}
		if lsa != "" {
			up["last_sync_at"] = lsa
		}
		_ = s.DB.Table("pipelines").Where("id = ?::uuid", pipelineID).Updates(up).Error
	}

	sm := strField(pipe, "sync_mode")
	if sm == "" {
		sm = "full"
	}
	target := "failed"
	if status == "completed" {
		if sm == "cdc" || sm == "log_based" || sm == "incremental" {
			target = "listing"
		} else {
			target = "idle"
		}
	} else {
		target = "failed"
	}

	if status == "completed" && (sm == "cdc" || sm == "log_based") && repMethod == "LOG_BASED" {
		var srcSchemaID string
		_ = s.DB.Raw(`SELECT source_schema_id::text FROM pipelines WHERE id = ?::uuid`, pipelineID).Scan(&srcSchemaID)
		var dsID string
		_ = s.DB.Raw(`SELECT data_source_id::text FROM pipeline_source_schemas WHERE id = ?::uuid`, srcSchemaID).Scan(&dsID)
		var connID string
		_ = s.DB.Raw(`SELECT id::text FROM data_source_connections WHERE data_source_id = ?::uuid LIMIT 1`, dsID).Scan(&connID)
		var existing sql.NullString
		_ = s.DB.Raw(`SELECT replication_slot_name FROM data_source_connections WHERE data_source_id = ?::uuid LIMIT 1`, dsID).Scan(&existing)
		if connID != "" && (!existing.Valid || existing.String == "") {
			slot := "mxf_" + strings.ReplaceAll(connID, "-", "")[:8]
			_ = s.DB.Exec(`UPDATE data_source_connections SET replication_slot_name = ? WHERE data_source_id = ?::uuid`, slot, dsID).Error
		}
	}

	totalProc := num(pipe["total_rows_processed"]) + rowsUpserted
	ts := int64(0)
	tf := int64(0)
	if status == "completed" {
		ts = 1
	} else {
		tf = 1
	}
	_ = s.DB.Table("pipelines").Where("id = ?::uuid", pipelineID).Updates(map[string]interface{}{
		"last_run_at":           time.Now(),
		"last_run_status":       runStatus,
		"status":                target,
		"total_rows_processed":  totalProc,
		"total_runs_successful": num(pipe["total_runs_successful"]) + int(ts),
		"total_runs_failed":     num(pipe["total_runs_failed"]) + int(tf),
		"last_sync_at":          time.Now(),
		"last_error":            san,
		"updated_at":            time.Now(),
	}).Error

	if status == "completed" && strField(pipe, "destination_schema_id") != "" {
		_ = s.DB.Exec(`UPDATE pipeline_destination_schemas SET last_synced_at = NOW(), updated_at = NOW() WHERE id = ?::uuid`,
			strField(pipe, "destination_schema_id")).Error
	}

	setFull := status == "completed" && (sm == "cdc" || sm == "log_based") && repMethod == "FULL_TABLE"
	if setFull {
		_ = s.DB.Exec(`UPDATE pipelines SET full_refresh_completed_at = NOW() WHERE id = ?::uuid`, pipelineID).Error
	}

	if s.Q != nil {
		orgID := strField(pipe, "organization_id")
		var er *string
		if san != "" {
			er = &san
		}
		s.Q.PublishStatus(c.Context(), pipelineID, orgID, target, &rowsUpserted, er)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"received": true})
}

func num(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return 0
	}
}
