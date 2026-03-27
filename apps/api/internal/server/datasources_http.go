package server

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/mantrixflow/go-api/pkg/response"
)

// ListDataSources returns organization data sources (excludes soft-deleted when column exists).
func (s *State) ListDataSources(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	var raw []byte
	q := `SELECT COALESCE(jsonb_agg(to_jsonb(ds) ORDER BY ds.created_at DESC), '[]'::jsonb)
		FROM data_sources ds
		WHERE ds.organization_id = ?::uuid AND (ds.deleted_at IS NULL)`
	if err := s.DB.Raw(q, orgID).Scan(&raw).Error; err != nil {
		return response.Error(c, http.StatusInternalServerError, "ERROR", err.Error())
	}
	var arr []interface{}
	_ = json.Unmarshal(raw, &arr)
	return response.List(c, arr, "OK", fiber.Map{"total": len(arr), "limit": len(arr), "offset": 0, "hasMore": false})
}

func (s *State) GetDataSource(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	id := c.Params("dataSourceId")
	var raw []byte
	if err := s.DB.Raw(`SELECT to_jsonb(ds) FROM data_sources ds WHERE ds.id = ?::uuid AND ds.organization_id = ?::uuid AND ds.deleted_at IS NULL`,
		id, orgID).Scan(&raw).Error; err != nil || len(raw) == 0 {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "Data source not found")
	}
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	return response.Success(c, m, "OK", http.StatusOK)
}

func (s *State) CreateDataSource(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
	}
	name := strField(body, "name")
	ct := strField(body, "connector_type")
	if ct == "" {
		ct = strField(body, "type")
	}
	if name == "" || ct == "" {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "name and connector_type (or type) required")
	}
	var id string
	if err := s.DB.Raw(`
		INSERT INTO data_sources (organization_id, name, connector_type, created_at, updated_at)
		VALUES (?::uuid, ?, ?, NOW(), NOW()) RETURNING id::text`,
		orgID, name, ct).Scan(&id).Error; err != nil || id == "" {
		return response.Error(c, http.StatusInternalServerError, "ERROR", "failed to create data source")
	}
	var raw []byte
	if err := s.DB.Raw(`SELECT to_jsonb(ds) FROM data_sources ds WHERE ds.id = ?::uuid AND ds.organization_id = ?::uuid`,
		id, orgID).Scan(&raw).Error; err != nil || len(raw) == 0 {
		return response.Error(c, http.StatusInternalServerError, "ERROR", "created but could not load")
	}
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	return response.Success(c, m, "Created", http.StatusCreated)
}

// PatchDataSource updates name / display metadata.
func (s *State) PatchDataSource(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	id := c.Params("dataSourceId")
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
	}
	if v, ok := body["name"].(string); ok && v != "" {
		res := s.DB.Exec(`UPDATE data_sources SET name = ?, updated_at = NOW() WHERE id = ?::uuid AND organization_id = ?::uuid AND deleted_at IS NULL`,
			v, id, orgID)
		if res.Error != nil || res.RowsAffected == 0 {
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "not found")
		}
	}
	return s.GetDataSource(c)
}

func (s *State) DeleteDataSource(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	id := c.Params("dataSourceId")
	res := s.DB.Exec(`UPDATE data_sources SET deleted_at = NOW(), updated_at = NOW() WHERE id = ?::uuid AND organization_id = ?::uuid AND deleted_at IS NULL`,
		id, orgID)
	if res.Error != nil || res.RowsAffected == 0 {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "not found")
	}
	return response.Success(c, fiber.Map{"deleted": true}, "OK", http.StatusOK)
}

// GetConnectionMasked returns decrypted-then-redacted connection config for UI.
func (s *State) GetConnectionMasked(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	dsID := c.Params("dataSourceId")
	ct, cfg, st, _, err := s.loadConnection(orgID, dsID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "connection not found")
	}
	dec, err := s.DecryptConnectionJSON(ct, cfg)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "ERROR", err.Error())
	}
	return response.Success(c, fiber.Map{
		"connection_type": ct,
		"status":          st,
		"config":          redactDecryptedConfig(dec),
	}, "OK", http.StatusOK)
}

// UpsertConnection encrypts and stores connection config for a data source.
func (s *State) UpsertConnection(c *fiber.Ctx) error {
	orgID := c.Params("organizationId")
	dsID := c.Params("dataSourceId")
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
	}
	connType := strField(body, "connection_type")
	if connType == "" {
		connType = strField(body, "connector_type")
	}
	if connType == "" {
		connType = strField(body, "type")
	}
	cfg, _ := body["config"].(map[string]interface{})
	if cfg == nil {
		cfg, _ = body["connection_config"].(map[string]interface{})
	}
	if connType == "" || cfg == nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "connection_type and config required")
	}
	var n int64
	_ = s.DB.Raw(`SELECT COUNT(*) FROM data_sources WHERE id = ?::uuid AND organization_id = ?::uuid AND deleted_at IS NULL`, dsID, orgID).Scan(&n)
	if n == 0 {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", "data source not found")
	}
	enc, err := s.EncryptConnectionJSON(connType, cfg)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	}
	raw, err := json.Marshal(enc)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "ERROR", err.Error())
	}
	res := s.DB.Exec(`
		UPDATE data_source_connections SET connection_type = ?, config = ?::jsonb, status = 'active', updated_at = NOW()
		WHERE data_source_id = ?::uuid`,
		connType, string(raw), dsID)
	if res.Error != nil {
		return response.Error(c, http.StatusInternalServerError, "ERROR", res.Error.Error())
	}
	if res.RowsAffected == 0 {
		ins := s.DB.Exec(`
			INSERT INTO data_source_connections (data_source_id, connection_type, config, status, created_at, updated_at)
			VALUES (?::uuid, ?, ?::jsonb, 'active', NOW(), NOW())`,
			dsID, connType, string(raw))
		if ins.Error != nil {
			return response.Error(c, http.StatusInternalServerError, "ERROR", ins.Error.Error())
		}
	}
	return s.GetConnectionMasked(c)
}
