package server

import (
	"fmt"
	"strings"

	"github.com/mantrixflow/go-api/internal/crypto"
)

func looksEncrypted(s string) bool {
	parts := strings.Split(s, ":")
	return len(parts) == 4 && len(parts[0]) > 20
}

// EncryptConnectionJSON mirrors DecryptConnectionJSON: encrypts password, SSL material, SSH key.
func (s *State) EncryptConnectionJSON(connectionType string, cfg map[string]interface{}) (map[string]interface{}, error) {
	t := strings.ToLower(connectionType)
	sqlTypes := map[string]bool{
		"postgres": true, "mysql": true, "mariadb": true, "mssql": true,
		"oracle": true, "sqlite": true, "cockroachdb": true, "postgresql": true,
	}
	if !sqlTypes[t] && t != "" {
		return cfg, nil
	}
	out := cloneMapIface(cfg)

	encStr := func(in string) (string, error) {
		if in == "" || looksEncrypted(in) {
			return in, nil
		}
		return crypto.Encrypt(s.Cfg.EncryptionMasterKey, in)
	}

	if p, ok := out["password"].(string); ok && p != "" {
		v, err := encStr(p)
		if err != nil {
			return nil, fmt.Errorf("password: %w", err)
		}
		out["password"] = v
	}
	if ssl, ok := out["ssl"].(map[string]interface{}); ok {
		for _, k := range []string{"ca_cert", "client_cert", "client_key"} {
			if ssl[k] == nil {
				continue
			}
			ps, ok := ssl[k].(string)
			if !ok {
				continue
			}
			v, err := encStr(ps)
			if err != nil {
				return nil, fmt.Errorf("ssl.%s: %w", k, err)
			}
			ssl[k] = v
		}
		out["ssl"] = ssl
	}
	if tun, ok := out["ssh_tunnel"].(map[string]interface{}); ok {
		if pk, ok := tun["private_key"].(string); ok && pk != "" {
			v, err := encStr(pk)
			if err != nil {
				return nil, fmt.Errorf("ssh_tunnel.private_key: %w", err)
			}
			tun["private_key"] = v
		}
		out["ssh_tunnel"] = tun
	}
	return out, nil
}

func redactDecryptedConfig(cfg map[string]interface{}) map[string]interface{} {
	out := cloneMapIface(cfg)
	if _, ok := out["password"].(string); ok {
		out["password"] = "***"
	}
	if ssl, ok := out["ssl"].(map[string]interface{}); ok {
		for _, k := range []string{"ca_cert", "client_cert", "client_key"} {
			if ssl[k] != nil {
				ssl[k] = "***"
			}
		}
		out["ssl"] = ssl
	}
	if tun, ok := out["ssh_tunnel"].(map[string]interface{}); ok {
		if tun["private_key"] != nil {
			tun["private_key"] = "***"
		}
		out["ssh_tunnel"] = tun
	}
	return out
}
