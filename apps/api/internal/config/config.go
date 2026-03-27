package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                         string
	DatabaseURL                  string
	DatabaseDirectURL            string
	SupabaseJWTSecret            string
	SupabaseURL                  string
	SupabaseAnonKey              string
	EncryptionMasterKey          string
	ETLPythonServiceURL          string
	ETLInternalToken             string
	CallbackToken                string
	InternalToken                string
	APIPublicURL                 string // e.g. https://api.example.com (callbacks; no trailing slash)
	AllowSourceDBMutationsForCDC bool
	PGMQPollIntervalMS           int
	PGMQParallelWorkers          int
	PGMQVTLongSec                int
	PGMQVTShortSec               int
	PGMQMaxDispatchRetries       int
	LogLevel                     string
	Environment                  string
}

func getenv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../app/.env") // optional shared monorepo env

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	jwtSecret := strings.TrimSpace(os.Getenv("SUPABASE_JWT_SECRET"))
	if jwtSecret == "" {
		jwtSecret = strings.TrimSpace(os.Getenv("JWT_SECRET"))
	}
	if jwtSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET (or JWT_SECRET) is required")
	}
	enc := strings.TrimSpace(os.Getenv("ENCRYPTION_MASTER_KEY"))
	if enc == "" || len(enc) < 32 {
		return nil, fmt.Errorf("ENCRYPTION_MASTER_KEY must be set (min 32 chars)")
	}
	etlURL := strings.TrimSpace(os.Getenv("ETL_PYTHON_SERVICE_URL"))
	if etlURL == "" {
		etlURL = strings.TrimSpace(os.Getenv("PYTHON_SERVICE_URL"))
	}
	if etlURL == "" {
		return nil, fmt.Errorf("ETL_PYTHON_SERVICE_URL is required")
	}
	pub := strings.TrimSpace(os.Getenv("API_PUBLIC_URL"))
	if pub == "" {
		pub = fmt.Sprintf("http://127.0.0.1:%s", getenv("PORT", "8080"))
	}

	pollMs, _ := strconv.Atoi(getenv("PGMQ_POLL_INTERVAL_MS", "2000"))
	parallel, _ := strconv.Atoi(getenv("PGMQ_PARALLEL_WORKERS", "50"))
	vtLong, _ := strconv.Atoi(getenv("PGMQ_VT_LONG_SEC", "14400"))
	vtShort, _ := strconv.Atoi(getenv("PGMQ_VT_SHORT_SEC", "300"))
	maxRetry, _ := strconv.Atoi(getenv("PGMQ_MAX_DISPATCH_RETRIES", "10"))

	rawCDC := strings.ToLower(strings.TrimSpace(os.Getenv("ALLOW_SOURCE_DB_MUTATIONS_FOR_CDC")))
	allowCDC := rawCDC == "1" || rawCDC == "true" || rawCDC == "yes" || rawCDC == "on"

	return &Config{
		Port:                         getenv("PORT", "8080"),
		DatabaseURL:                  dbURL,
		DatabaseDirectURL:            strings.TrimSpace(os.Getenv("DATABASE_DIRECT_URL")),
		SupabaseJWTSecret:            jwtSecret,
		SupabaseURL:                  strings.TrimSpace(os.Getenv("SUPABASE_URL")),
		SupabaseAnonKey:              strings.TrimSpace(os.Getenv("SUPABASE_ANON_KEY")),
		EncryptionMasterKey:          enc,
		ETLPythonServiceURL:          strings.TrimRight(etlURL, "/"),
		ETLInternalToken:             firstNonEmpty(os.Getenv("ETL_INTERNAL_TOKEN"), os.Getenv("INTERNAL_TOKEN"), os.Getenv("SUPABASE_SERVICE_ROLE_KEY")),
		CallbackToken:                firstNonEmpty(os.Getenv("CALLBACK_TOKEN"), os.Getenv("INTERNAL_TOKEN"), os.Getenv("ETL_INTERNAL_TOKEN")),
		InternalToken:                strings.TrimSpace(os.Getenv("INTERNAL_TOKEN")),
		APIPublicURL:                 strings.TrimRight(pub, "/"),
		AllowSourceDBMutationsForCDC: allowCDC,
		PGMQPollIntervalMS:           pollMs,
		PGMQParallelWorkers:          parallel,
		PGMQVTLongSec:                vtLong,
		PGMQVTShortSec:               vtShort,
		PGMQMaxDispatchRetries:       maxRetry,
		LogLevel:                     getenv("LOG_LEVEL", "info"),
		Environment:                  getenv("ENVIRONMENT", "development"),
	}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func (c *Config) SessionDatabaseURL() string {
	if c.DatabaseDirectURL != "" {
		return c.DatabaseDirectURL
	}
	u := c.DatabaseURL
	if strings.Contains(u, ":6543") {
		return strings.Replace(u, ":6543", ":5432", 1)
	}
	return u
}
