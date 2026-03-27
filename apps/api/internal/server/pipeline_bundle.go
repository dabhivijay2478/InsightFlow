package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (s *State) loadPipelineBundle(tx *gorm.DB, pipelineID string) (pipe, src, dest map[string]interface{}, err error) {
	if tx == nil {
		tx = s.DB
	}
	var pj, sj, dj datatypes.JSON
	row := tx.Raw(`
		SELECT to_jsonb(p), to_jsonb(ss), to_jsonb(dd)
		FROM pipelines p
		INNER JOIN pipeline_source_schemas ss ON ss.id = p.source_schema_id
		INNER JOIN pipeline_destination_schemas dd ON dd.id = p.destination_schema_id
		WHERE p.id = ?::uuid AND p.deleted_at IS NULL
		LIMIT 1`, pipelineID).Row()
	if row.Err() != nil {
		return nil, nil, nil, row.Err()
	}
	if err := row.Scan(&pj, &sj, &dj); err != nil {
		return nil, nil, nil, err
	}
	if len(pj) == 0 {
		return nil, nil, nil, fmt.Errorf("pipeline not found")
	}
	_ = json.Unmarshal(pj, &pipe)
	_ = json.Unmarshal(sj, &src)
	_ = json.Unmarshal(dj, &dest)
	return pipe, src, dest, nil
}

func graphNodes(pipe map[string]interface{}) []interface{} {
	g, ok := pipe["pipeline_graph"].(map[string]interface{})
	if !ok || g == nil {
		return nil
	}
	n, _ := g["nodes"].([]interface{})
	return n
}

func effectiveDestinationSchema(pipe, dest map[string]interface{}) map[string]interface{} {
	out := cloneMapIface(dest)
	nodes := graphNodes(pipe)
	var graphDest map[string]interface{}
	dsID, _ := dest["data_source_id"].(string)
	for _, n := range nodes {
		node, ok := n.(map[string]interface{})
		if !ok || node["type"] != "destination" {
			continue
		}
		data, _ := node["data"].(map[string]interface{})
		if data == nil {
			continue
		}
		if cid, ok := data["connection_id"].(string); ok && cid != "" && cid == dsID {
			graphDest = data
			break
		}
		if graphDest == nil {
			graphDest = data
		}
	}
	if graphDest == nil {
		return out
	}
	if v, ok := graphDest["dest_schema"].(string); ok && v != "" {
		out["destination_schema"] = v
	}
	if v, ok := graphDest["dest_table"].(string); ok && v != "" {
		out["destination_table"] = v
	}
	if v, ok := graphDest["emit_method"].(string); ok && v != "" {
		out["write_mode"] = v
	}
	if cid, ok := graphDest["connection_id"].(string); ok && cid != "" {
		out["data_source_id"] = cid
	}
	return out
}

func extractColumnMap(pipe map[string]interface{}) map[string]string {
	nodes := graphNodes(pipe)
	for _, n := range nodes {
		node, ok := n.(map[string]interface{})
		if !ok || node["type"] != "transform" {
			continue
		}
		data, _ := node["data"].(map[string]interface{})
		if data == nil {
			continue
		}
		if cm, ok := data["column_map"].(map[string]interface{}); ok && len(cm) > 0 {
			out := make(map[string]string)
			for k, v := range cm {
				if s, ok := v.(string); ok {
					out[k] = s
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	if tr, ok := pipe["transformations"].([]interface{}); ok {
		out := make(map[string]string)
		for _, t := range tr {
			m, ok := t.(map[string]interface{})
			if !ok {
				continue
			}
			if m["transformType"] != "rename" {
				continue
			}
			sc, _ := m["sourceColumn"].(string)
			dc, _ := m["destinationColumn"].(string)
			if sc != "" && dc != "" {
				out[dc] = sc
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

func extractTransformScript(pipe, dest map[string]interface{}) string {
	nodes := graphNodes(pipe)
	for _, n := range nodes {
		node, ok := n.(map[string]interface{})
		if !ok || node["type"] != "transform" {
			continue
		}
		data, _ := node["data"].(map[string]interface{})
		if data == nil {
			continue
		}
		if sc, ok := data["transform_script"].(string); ok && strings.TrimSpace(sc) != "" {
			return sc
		}
	}
	if sc, ok := dest["transform_script"].(string); ok {
		return sc
	}
	return ""
}

func replicationMethodFromSyncMode(syncMode string) string {
	switch strings.ToLower(syncMode) {
	case "cdc", "log_based":
		return "LOG_BASED"
	case "incremental":
		return "INCREMENTAL"
	default:
		return "FULL_TABLE"
	}
}

func registrySourceType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "postgresql", "postgres", "pg":
		return "postgres"
	default:
		return strings.ToLower(strings.TrimSpace(t))
	}
}

func strField(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func getStr(m map[string]interface{}, k string) string {
	return strField(m, k)
}
