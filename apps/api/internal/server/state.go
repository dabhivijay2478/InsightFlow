package server

import (
	"github.com/mantrixflow/go-api/internal/config"
	"github.com/mantrixflow/go-api/internal/etl"
	"github.com/mantrixflow/go-api/internal/queue"
	"gorm.io/gorm"
)

// State holds shared dependencies for HTTP handlers and the worker.
type State struct {
	Cfg *config.Config
	DB  *gorm.DB
	Q   *queue.Service
	ETL *etl.Client
}

func NewState(cfg *config.Config, db *gorm.DB, q *queue.Service, etlClient *etl.Client) *State {
	return &State{Cfg: cfg, DB: db, Q: q, ETL: etlClient}
}
