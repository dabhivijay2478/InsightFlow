package server

import (
	"fmt"
	"strings"

	"github.com/mantrixflow/go-api/internal/crypto"
)

func (s *State) decryptField(val interface{}) (interface{}, error) {
	str, ok := val.(string)
	if !ok || str == "" {
		return val, nil
	}
	parts := strings.Split(str, ":")
	if len(parts) != 4 {
		return str, nil
	}
	out, err := crypto.Decrypt(s.Cfg.EncryptionMasterKey, str)
	if err != nil && strings.Contains(err.Error(), "invalid") {
		return str, nil
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DecryptConnectionJSON mutates a config map like Nest decryptConfig for SQL connectors.
func (s *State) DecryptConnectionJSON(connectionType string, cfg map[string]interface{}) (map[string]interface{}, error) {
	t := strings.ToLower(connectionType)
	sqlTypes := map[string]bool{
		"postgres": true, "mysql": true, "mariadb": true, "mssql": true,
		"oracle": true, "sqlite": true, "cockroachdb": true, "postgresql": true,
	}
	if !sqlTypes[t] && t != "" {
		return cfg, nil
	}
	out := cloneMapIface(cfg)
	if p, ok := out["password"].(string); ok && p != "" {
		v, err := s.decryptField(p)
		if err != nil {
			return nil, fmt.Errorf("password: %w", err)
		}
		out["password"] = v
	}
	if ssl, ok := out["ssl"].(map[string]interface{}); ok {
		for _, k := range []string{"ca_cert", "client_cert", "client_key"} {
			if ssl[k] != nil {
				v, err := s.decryptField(ssl[k])
				if err != nil {
					return nil, err
				}
				ssl[k] = v
			}
		}
		out["ssl"] = ssl
	}
	if tun, ok := out["ssh_tunnel"].(map[string]interface{}); ok {
		if tun["private_key"] != nil {
			v, err := s.decryptField(tun["private_key"])
			if err != nil {
				return nil, err
			}
			tun["private_key"] = v
		}
		out["ssh_tunnel"] = tun
	}
	return out, nil
}

func cloneMapIface(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{}, len(m))
	for k, v := range m {
		switch t := v.(type) {
		case map[string]interface{}:
			o[k] = cloneMapIface(t)
		default:
			o[k] = v
		}
	}
	return o
}
