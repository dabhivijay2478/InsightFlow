package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mantrixflow/go-api/internal/connectorsdata"
	"github.com/mantrixflow/go-api/pkg/response"
)

func (s *State) ListPipelines(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	var raw []byte
	q := `SELECT COALESCE(jsonb_agg(x.j ORDER BY x.created_at DESC), '[]'::jsonb) FROM (
		SELECT to_jsonb(p) AS j, p.created_at FROM pipelines p
		WHERE p.organization_id = ?::uuid AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC LIMIT ? OFFSET ?
	) x`
	if err := s.DB.Raw(q, orgID, limit, offset).Scan(&raw).Error; err != nil {
		return response.Error(c, http.StatusInternalServerError, "ERROR", err.Error())
	}
	var arr []interface{}
	_ = json.Unmarshal(raw, &arr)
	meta := fiber.Map{"total": len(arr), "limit": len(arr), "offset": 0, "hasMore": false}
	return response.List(c, arr, "OK", meta)
}

func (s *State) GetPipeline(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	id := c.Params("id")
	var raw []byte
	if err := s.DB.Raw(`SELECT to_jsonb(p) FROM pipelines p WHERE p.id = ?::uuid AND p.organization_id = ?::uuid AND p.deleted_at IS NULL`,
		id, orgID).Scan(&raw).Error; err != nil || len(raw) == 0 {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "Pipeline not found")
	}
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	return response.Success(c, m, "OK", http.StatusOK)
}

func (s *State) GetPipelineFull(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	id := c.Params("id")
	pipe, src, dest, err := s.loadPipelineBundle(nil, id)
	if err != nil || strField(pipe, "organization_id") != orgID {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "Pipeline not found")
	}
	return response.Success(c, fiber.Map{
		"pipeline": pipe, "sourceSchema": src, "destinationSchema": dest,
	}, "OK", http.StatusOK)
}

func (s *State) CreatePipeline(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	uid := UserID(c)
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
	}
	name := strField(body, "name")
	srcSch := strField(body, "sourceSchemaId")
	dstSch := strField(body, "destinationSchemaId")
	if name == "" || srcSch == "" || dstSch == "" {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "name, sourceSchemaId, destinationSchemaId required")
	}
	var cnt int64
	s.DB.Raw(`SELECT COUNT(*) FROM pipelines WHERE organization_id = ?::uuid AND name = ? AND deleted_at IS NULL`, orgID, name).Scan(&cnt)
	if cnt > 0 {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Pipeline name already exists")
	}
	desc := strField(body, "description")
	syncMode := strField(body, "syncMode")
	if syncMode == "" {
		syncMode = "full"
	}
	incCol := strField(body, "incrementalColumn")
	trans, _ := json.Marshal(body["transformations"])
	pg, _ := json.Marshal(body["pipelineGraph"])
	bvm := strField(body, "builderViewMode")
	if bvm == "" {
		bvm = "card"
	}
	type idRow struct{ ID string }
	var out idRow
	tb := string(trans)
	if tb == "" || tb == "null" {
		tb = "null"
	}
	pgb := string(pg)
	if pgb == "" || pgb == "null" {
		pgb = "null"
	}
	err := s.DB.Raw(`
		INSERT INTO pipelines (
			organization_id, created_by, name, description,
			source_schema_id, destination_schema_id,
			transformations, sync_mode, incremental_column,
			sync_frequency, status, schedule_type,
			pipeline_graph, builder_view_mode,
			created_at, updated_at
		) VALUES (
			?::uuid, ?::uuid, ?, ?::text,
			?::uuid, ?::uuid,
			?::jsonb, ?, ?::text,
			'manual', 'idle', 'none',
			?::jsonb, ?,
			NOW(), NOW()
		) RETURNING id::text`,
		orgID, uid, name, nullStr(desc),
		srcSch, dstSch,
		tb, syncMode, nullStr(incCol),
		pgb, bvm,
	).Scan(&out).Error
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	}
	var raw []byte
	_ = s.DB.Raw(`SELECT to_jsonb(p) FROM pipelines p WHERE p.id = ?::uuid`, out.ID).Scan(&raw).Error
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	return response.Success(c, m, "Pipeline created", http.StatusCreated)
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (s *State) PatchPipeline(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	id := c.Params("id")
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
	}
	up := map[string]interface{}{"updated_at": time.Now()}
	if v, ok := body["name"].(string); ok {
		up["name"] = v
	}
	if v, ok := body["description"].(string); ok {
		up["description"] = v
	}
	if v, ok := body["syncMode"].(string); ok {
		up["sync_mode"] = v
	}
	if v, ok := body["incrementalColumn"].(string); ok {
		up["incremental_column"] = v
	}
	if v, ok := body["transformations"]; ok {
		b, _ := json.Marshal(v)
		up["transformations"] = string(b)
	}
	if v, ok := body["pipelineGraph"]; ok {
		b, _ := json.Marshal(v)
		up["pipeline_graph"] = string(b)
	}
	if v, ok := body["scheduleType"].(string); ok {
		up["schedule_type"] = v
	}
	if v, ok := body["scheduleValue"].(string); ok {
		up["schedule_value"] = v
	}
	res := s.DB.Table("pipelines").Where("id = ?::uuid AND organization_id = ?::uuid AND deleted_at IS NULL", id, orgID).Updates(up)
	if res.Error != nil || res.RowsAffected == 0 {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "Pipeline not found")
	}
	return s.GetPipeline(c)
}

func (s *State) DeletePipeline(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	id := c.Params("id")
	res := s.DB.Exec(`UPDATE pipelines SET deleted_at = NOW(), updated_at = NOW() WHERE id = ?::uuid AND organization_id = ?::uuid AND deleted_at IS NULL`, id, orgID)
	if res.Error != nil || res.RowsAffected == 0 {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "Pipeline not found")
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"meta": fiber.Map{
			"statusCode": http.StatusOK, "message": "Deleted", "status": "success",
			"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		},
		"deletedId": id,
	})
}

func (s *State) ValidatePipeline(c *fiber.Ctx) error {
	id := c.Params("id")
	_, src, dest, err := s.loadPipelineBundle(nil, id)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "not found")
	}
	errs := []string{}
	if strField(src, "data_source_id") == "" {
		errs = append(errs, "Source schema must have a data source")
	}
	if strField(src, "source_table") == "" && strField(src, "source_query") == "" {
		errs = append(errs, "Source schema must have a source table or query")
	}
	if strField(dest, "data_source_id") == "" {
		errs = append(errs, "Destination schema must have a data source")
	}
	if strField(dest, "destination_table") == "" {
		errs = append(errs, "Destination schema must have a destination table")
	}
	if strField(dest, "write_mode") == "upsert" {
		if dest["upsert_key"] == nil {
			errs = append(errs, "Upsert mode requires upsert key columns")
		}
	}
	return response.Success(c, fiber.Map{"valid": len(errs) == 0, "errors": errs, "warnings": []string{}}, "OK", http.StatusOK)
}

func (s *State) DryRunPipeline(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	id := c.Params("id")
	uid := UserID(c)
	pipe, src, _, err := s.loadPipelineBundle(nil, id)
	if err != nil || strField(pipe, "organization_id") != orgID {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "not found")
	}
	_ = uid
	srcDS := strField(src, "data_source_id")
	ct, _, _, _, err := s.loadConnection(orgID, srcDS)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "no connection")
	}
	cfg, err := s.decryptedConnMap(orgID, srcDS, ct)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "ERROR", err.Error())
	}
	srcSchema := strField(src, "source_schema")
	srcTable := strField(src, "source_table")
	stream := srcTable
	if srcSchema != "" && srcTable != "" {
		stream = srcSchema + "-" + srcTable
	}
	body := map[string]interface{}{
		"source_type":   registrySourceType(ct),
		"source_config": cfg,
		"source_stream": stream,
		"limit":         10,
	}
	prev, err := s.ETL.Preview(body)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	}
	rec, _ := prev["records"].([]interface{})
	total := num(prev["total"])
	return response.Success(c, fiber.Map{
		"wouldWrite": total, "sourceRowCount": total, "sampleRows": rec,
		"transformedSample": rec, "errors": []string{}, "appliedMappings": []interface{}{},
	}, "OK", http.StatusOK)
}

func (s *State) ListPipelineRuns(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	pid := c.Params("id")
	var raw []byte
	_ = s.DB.Raw(`
		SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY r.created_at DESC), '[]'::jsonb)
		FROM pipeline_runs r
		INNER JOIN pipelines p ON p.id = r.pipeline_id AND p.organization_id = ?::uuid
		WHERE r.pipeline_id = ?::uuid LIMIT 100`,
		orgID, pid).Scan(&raw).Error
	var arr []interface{}
	_ = json.Unmarshal(raw, &arr)
	return response.List(c, arr, "OK", fiber.Map{"total": len(arr), "limit": 100, "offset": 0, "hasMore": false})
}

func (s *State) GetPipelineRun(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	rid := c.Params("runId")
	var raw []byte
	if err := s.DB.Raw(`
		SELECT to_jsonb(r) FROM pipeline_runs r
		INNER JOIN pipelines p ON p.id = r.pipeline_id AND p.organization_id = ?::uuid
		WHERE r.id = ?::uuid`, orgID, rid).Scan(&raw).Error; err != nil || len(raw) == 0 {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "not found")
	}
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	return response.Success(c, m, "OK", http.StatusOK)
}

func (s *State) PausePipeline(c *fiber.Ctx) error {
	return s.setPipelineStatus(c, "paused")
}

func (s *State) ResumePipeline(c *fiber.Ctx) error {
	return s.setPipelineStatus(c, "idle")
}

func (s *State) setPipelineStatus(c *fiber.Ctx, st string) error {
	orgID := c.Params("organizationId")
	id := c.Params("id")
	res := s.DB.Exec(`UPDATE pipelines SET status = ?, updated_at = NOW() WHERE id = ?::uuid AND organization_id = ?::uuid AND deleted_at IS NULL`, st, id, orgID)
	if res.Error != nil || res.RowsAffected == 0 {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "not found")
	}
	return s.GetPipeline(c)
}

func (s *State) GetSyncState(c *fiber.Ctx) error {
	id := c.Params("id")
	var chk []byte
	_ = s.DB.Raw(`SELECT checkpoint FROM pipelines WHERE id = ?::uuid`, id).Scan(&chk).Error
	var cp interface{}
	_ = json.Unmarshal(chk, &cp)
	return response.Success(c, fiber.Map{"checkpoint": cp}, "OK", http.StatusOK)
}

func (s *State) GetPipelineStats(c *fiber.Ctx) error {
	return response.Success(c, fiber.Map{"totalRuns": 0, "successRate": 0}, "OK", http.StatusOK)
}

func (s *State) GetScheduleInfo(c *fiber.Ctx) error {
	return s.GetPipeline(c)
}

func (s *State) GetCdcStatus(c *fiber.Ctx) error {
	return response.Success(c, fiber.Map{"verified": false}, "OK", http.StatusOK)
}

func (s *State) CancelRun(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	rid := c.Params("runId")
	_ = s.DB.Exec(`
		UPDATE pipeline_runs r SET status = 'cancelled', job_state = 'failed', updated_at = NOW()
		FROM pipelines p
		WHERE r.id = ?::uuid AND r.pipeline_id = p.id AND p.organization_id = ?::uuid`,
		rid, orgID).Error
	return s.GetPipelineRun(c)
}

func (s *State) ProxyETLTestConnection(c *fiber.Ctx) error {
	var body map[string]interface{}
	_ = c.BodyParser(&body)
	if body == nil {
		body = map[string]interface{}{}
	}
	st := strField(body, "source_type")
	if st == "" {
		st = strField(body, "sourceType")
	}
	cfg, _ := body["source_config"].(map[string]interface{})
	if cfg == nil {
		cfg, _ = body["connection_config"].(map[string]interface{})
	}
	out, err := s.ETL.TestConnection(map[string]interface{}{
		"source_type":       registrySourceType(st),
		"connection_config": cfg,
	})
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	}
	return response.Success(c, out, "OK", http.StatusOK)
}

func (s *State) ProxyETLDiscover(c *fiber.Ctx) error {
	var body map[string]interface{}
	_ = c.BodyParser(&body)
	if body == nil {
		body = map[string]interface{}{}
	}
	out, err := s.ETL.Discover(body)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	}
	return response.Success(c, out, "OK", http.StatusOK)
}

func (s *State) ProxyETLPreview(c *fiber.Ctx) error {
	var body map[string]interface{}
	_ = c.BodyParser(&body)
	if body == nil {
		body = map[string]interface{}{}
	}
	out, err := s.ETL.Preview(body)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	}
	return response.Success(c, out, "OK", http.StatusOK)
}

func (s *State) ProxyETLStub(c *fiber.Ctx) error {
	return response.Success(c, fiber.Map{"ok": true}, "Stub", http.StatusOK)
}

func (s *State) ConnectorsMetadata(c *fiber.Ctx) error {
	data := connectorsResponse()
	return response.Success(c, data, "OK", http.StatusOK)
}

func (s *State) ConnectorsHealth(c *fiber.Ctx) error {
	out, err := s.ETL.GetJSON("/health")
	if err != nil {
		return response.Success(c, fiber.Map{"status": "degraded", "detail": err.Error()}, "OK", http.StatusOK)
	}
	return response.Success(c, out, "OK", http.StatusOK)
}

func (s *State) ConnectorsCdcSetup(c *fiber.Ctx) error {
	t := c.Params("sourceType")
	return response.Success(c, fiber.Map{
		"source_type":   t,
		"cdc_supported": strings.EqualFold(t, "postgres") || strings.EqualFold(t, "postgresql"),
		"setup": fiber.Map{
			"title": "CDC prerequisites",
			"steps": []string{"Enable logical replication on the source", "Grant replication to the DB user"},
		},
	}, "OK", http.StatusOK)
}

func (s *State) ConnectorsTestConnection(c *fiber.Ctx) error {
	var body map[string]interface{}
	_ = c.BodyParser(&body)
	if body == nil {
		body = map[string]interface{}{}
	}
	ct := strField(body, "connectionType")
	if ct == "" {
		ct = strField(body, "connection_type")
	}
	cfg, _ := body["config"].(map[string]interface{})
	out, err := s.ETL.TestConnection(map[string]interface{}{
		"source_type":       registrySourceType(ct),
		"connection_config": cfg,
	})
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	}
	return response.Success(c, out, "OK", http.StatusOK)
}

func connectorsResponse() map[string]interface{} {
	var m map[string]interface{}
	_ = json.Unmarshal(connectorsdata.ConnectorsJSON, &m)
	return m
}
