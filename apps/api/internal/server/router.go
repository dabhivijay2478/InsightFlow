package server

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mantrixflow/go-api/pkg/response"
	"github.com/rs/zerolog/log"
)

// RegisterRoutes mounts /api/v1 and supporting paths on app.
func RegisterRoutes(s *State, app *fiber.App) {
	app.Get("/health", health)
	app.Head("/health", health)

	v1 := app.Group("/api/v1")
	v1.Get("/health", health)
	v1.Head("/health", health)

	internal := v1.Group("/internal", s.InternalToken())
	internal.Post("/etl-callback", s.EtlCallback)
	internal.Get("/checkpoint/:pipelineId", s.GetCheckpoint)
	internal.Post("/connections/resolve", s.ResolveConnections)
	internal.Post("/process-etl-jobs", s.ProcessEtlJobsStub)

	v1.Get("/connectors", s.ConnectorsMetadata)
	v1.Get("/connectors/health", s.ConnectorsHealth)

	authed := v1.Group("", s.AuthJWT())
	authed.Post("/connectors/test-connection", s.ConnectorsTestConnection)
	authed.Get("/connectors/cdc-setup/:sourceType", s.ConnectorsCdcSetup)

	org := authed.Group("/organizations/:organizationId", s.RequireOrgMember())
	org.Get("/dashboard/summary", s.StubOK)
	org.Get("/activity", s.StubOK)
	org.Get("/onboarding/status", s.StubOK)
	org.Post("/search/query", s.StubOK)

	ds := org.Group("/data-sources")
	ds.Get("/", s.ListDataSources)
	ds.Post("/", s.CreateDataSource)
	d1 := ds.Group("/:dataSourceId")
	d1.Get("", s.GetDataSource)
	d1.Patch("", s.PatchDataSource)
	d1.Delete("", s.DeleteDataSource)
	d1.Get("/connection", s.GetConnectionMasked)
	d1.Put("/connection", s.UpsertConnection)

	etlx := org.Group("/etl")
	etlx.Post("/test-connection", s.ProxyETLTestConnection)
	etlx.Post("/discover", s.ProxyETLDiscover)
	etlx.Post("/preview", s.ProxyETLPreview)
	etlx.Post("/stub", s.ProxyETLStub)

	pipes := org.Group("/pipelines")
	pipes.Get("/", s.ListPipelines)
	pipes.Post("/", s.CreatePipeline)
	p1 := pipes.Group("/:id")
	p1.Get("", s.GetPipeline)
	p1.Get("/full", s.GetPipelineFull)
	p1.Patch("", s.PatchPipeline)
	p1.Delete("", s.DeletePipeline)
	p1.Post("/validate", s.ValidatePipeline)
	p1.Post("/dry-run", s.DryRunPipeline)
	p1.Post("/run", s.RunPipelineHTTP)
	p1.Get("/runs", s.ListPipelineRuns)
	p1.Get("/runs/:runId", s.GetPipelineRun)
	p1.Post("/runs/:runId/cancel", s.CancelRun)
	p1.Get("/sync-state", s.GetSyncState)
	p1.Get("/stats", s.GetPipelineStats)
	p1.Get("/schedule", s.GetScheduleInfo)
	p1.Get("/cdc-status", s.GetCdcStatus)
	p1.Post("/pause", s.PausePipeline)
	p1.Post("/resume", s.ResumePipeline)
}

func health(c *fiber.Ctx) error {
	return response.Success(c, fiber.Map{"ok": true, "service": "api"}, "OK", http.StatusOK)
}

// StubOK returns an empty success payload for routes deferred from Nest parity.
func (s *State) StubOK(c *fiber.Ctx) error {
	return response.Success(c, fiber.Map{}, "OK", http.StatusOK)
}

// RequestLogger emits one line per request (zerolog).
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		t0 := time.Now()
		err := c.Next()
		path := c.Route().Path
		if path == "" {
			path = c.Path()
		}
		log.Info().
			Str("method", c.Method()).
			Str("path", path).
			Int("status", c.Response().StatusCode()).
			Dur("dur", time.Since(t0)).
			Msg("http")
		return err
	}
}
