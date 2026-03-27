package etl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mantrixflow/go-api/internal/config"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.ETLPythonServiceURL,
		token:   cfg.ETLInternalToken,
		http: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) headers() http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	if c.token != "" {
		h.Set("X-ETL-Token", c.token)
	}
	return h
}

func (c *Client) PostJSON(path string, body interface{}, timeout time.Duration) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header = c.headers()
	client := c.http
	if timeout > 0 {
		client = &http.Client{Timeout: timeout}
	}
	return client.Do(req)
}

func (c *Client) TestConnection(body interface{}) (map[string]interface{}, error) {
	resp, err := c.PostJSON("/test-connection", body, 60*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("etl test-connection: %s: %s", resp.Status, string(raw))
	}
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	return m, nil
}

func (c *Client) Discover(body interface{}) (map[string]interface{}, error) {
	resp, err := c.PostJSON("/discover", body, 120*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("etl discover: %s: %s", resp.Status, string(raw))
	}
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	return m, nil
}

func (c *Client) Preview(body interface{}) (map[string]interface{}, error) {
	resp, err := c.PostJSON("/preview", body, 120*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("etl preview: %s: %s", resp.Status, string(raw))
	}
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	return m, nil
}

func (c *Client) IntrospectTable(body interface{}) (map[string]interface{}, error) {
	resp, err := c.PostJSON("/introspect-table", body, 120*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("etl introspect: %s: %s", resp.Status, string(raw))
	}
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	return m, nil
}

// RunSync POST /sync; returns retry true if 503; error on other failures.
func (c *Client) RunSync(callbackBaseURL string, payload map[string]interface{}) (retry bool, err error) {
	payload = cloneMap(payload)
	payload["nestjs_callback_url"] = strings.TrimRight(callbackBaseURL, "/") + "/api/v1/internal/etl-callback"
	payload["nestjs_state_url"] = strings.TrimRight(callbackBaseURL, "/") + "/api/v1/internal/checkpoint"

	b, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/sync", bytes.NewReader(b))
	if err != nil {
		return false, err
	}
	req.Header = c.headers()
	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 503 {
		return true, nil
	}
	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("sync http %d: %s", resp.StatusCode, string(raw))
	}
	return false, nil
}

func cloneMap(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (c *Client) GetJSON(path string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header = c.headers()
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("etl get %s: %s", path, string(raw))
	}
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	return m, nil
}
