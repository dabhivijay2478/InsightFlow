package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mantrixflow/go-api/internal/queue"
	"github.com/mantrixflow/go-api/pkg/response"
)

func (s *State) RunPipelineHTTP(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	pipelineID := c.Params("id")
	uid := UserID(c)

	pipe, src, dest, err := s.loadPipelineBundle(nil, pipelineID)
	if err != nil || pipe == nil {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "Pipeline not found")
	}
	if strField(pipe, "organization_id") != orgID {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "Pipeline does not belong to organization")
	}
	if strField(pipe, "status") == "paused" {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Pipeline is paused. Resume it before running.")
	}

	sm := strField(pipe, "sync_mode")
	isCDC := sm == "cdc" || sm == "log_based"
	if isCDC && !s.Cfg.AllowSourceDBMutationsForCDC {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "CDC and LOG_BASED sync are disabled by policy because they can alter the client source database (for example, replication slots).")
	}
	needsInitial := isCDC && strField(pipe, "full_refresh_completed_at") == ""
	if isCDC && !needsInitial {
		srcDS := strField(src, "data_source_id")
		var cdc []byte
		_ = s.DB.Raw(`SELECT cdc_prerequisites_status FROM data_source_connections WHERE data_source_id = ?::uuid LIMIT 1`, srcDS).Scan(&cdc)
		overall := ""
		if len(cdc) > 0 {
			var m map[string]interface{}
			_ = json.Unmarshal(cdc, &m)
			if v, ok := m["overall"].(string); ok {
				overall = v
			}
		}
		if overall != "verified" {
			return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Complete the Log-Based Sync setup for this source connection before running.")
		}
	}

	destEff := effectiveDestinationSchema(pipe, dest)
	srcDS := strField(src, "data_source_id")
	dstDS := strField(destEff, "data_source_id")
	sc := s.loadConnRow(orgID, srcDS)
	dc := s.loadConnRow(orgID, dstDS)
	if sc == nil || dc == nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Cannot run pipeline: source or destination data source has no connection.")
	}
	if sc.status == "error" || dc.status == "error" {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Cannot run pipeline: source or destination connection has errors.")
	}

	type runIDRow struct{ ID string }
	var rr runIDRow
	insErr := s.DB.Raw(`
		INSERT INTO pipeline_runs (pipeline_id, organization_id, status, job_state, trigger_type, triggered_by, started_at, created_at, updated_at)
		VALUES (?::uuid, ?::uuid, 'pending', 'pending', 'manual', ?::uuid, NOW(), NOW(), NOW()) RETURNING id::text`,
		pipelineID, orgID, uid).Scan(&rr).Error
	if insErr != nil || rr.ID == "" {
		return response.Error(c, http.StatusInternalServerError, "ERROR", "failed to create pipeline run")
	}
	runID := rr.ID

	_ = s.DB.Table("pipelines").Where("id = ?::uuid", pipelineID).Updates(map[string]interface{}{
		"last_run_status": "running",
		"last_run_at":     time.Now(),
		"updated_at":      time.Now(),
	})

	ctx := context.Background()
	if s.Q != nil {
		isInc := sm == "incremental"
		if isCDC && needsInitial {
			_ = queue.EnqueueFullSync(ctx, s.Q, queue.FullSyncData{
				PipelineID: pipelineID, RunID: runID, OrganizationID: orgID, UserID: uid, TriggerType: "manual",
			})
		} else if isCDC || isInc {
			_ = queue.EnqueueIncrementalSync(ctx, s.Q, queue.IncrementalSyncData{
				PipelineID: pipelineID, RunID: runID, OrganizationID: orgID, UserID: uid, TriggerType: "manual",
			})
		} else {
			_ = queue.EnqueueFullSync(ctx, s.Q, queue.FullSyncData{
				PipelineID: pipelineID, RunID: runID, OrganizationID: orgID, UserID: uid, TriggerType: "manual",
			})
		}
		s.Q.PublishStatus(ctx, pipelineID, orgID, "pending", intPtr(0), nil)
	}

	raw := []byte{}
	_ = s.DB.Raw(`SELECT to_jsonb(r) FROM pipeline_runs r WHERE r.id = ?::uuid`, runID).Scan(&raw).Error
	var out map[string]interface{}
	_ = json.Unmarshal(raw, &out)
	return response.Success(c, out, "Pipeline run started", http.StatusCreated)
}
