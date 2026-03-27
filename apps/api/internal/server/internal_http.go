package server

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/mantrixflow/go-api/pkg/response"
)

type resolveDTO struct {
	OrganizationID string `json:"organization_id"`
	SourceConnID   string `json:"source_conn_id"`
	DestConnID     string `json:"dest_conn_id"`
}

func (s *State) ResolveConnections(c *fiber.Ctx) error {
	var body resolveDTO
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
	}
	srcDS := body.SourceConnID
	dstDS := body.DestConnID
	srcT := ""
	dstT := ""
	var srcCfg, dstCfg map[string]interface{}
	if srcDS != "" {
		var err error
		srcT, _, _, _, err = s.loadConnection(body.OrganizationID, srcDS)
		if err == nil {
			srcCfg, err = s.decryptedConnMap(body.OrganizationID, srcDS, srcT)
			if err != nil {
				srcCfg = nil
			}
		}
	}
	if dstDS != "" {
		var err error
		dstT, _, _, _, err = s.loadConnection(body.OrganizationID, dstDS)
		if err == nil {
			dstCfg, err = s.decryptedConnMap(body.OrganizationID, dstDS, dstT)
			if err != nil {
				dstCfg = nil
			}
		}
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"source": fiber.Map{"type": registrySourceType(srcT), "config": srcCfg},
		"dest":   fiber.Map{"type": registrySourceType(dstT), "config": dstCfg},
	})
}

func (s *State) GetCheckpoint(c *fiber.Ctx) error {
	pid := c.Params("pipelineId")
	var chk []byte
	_ = s.DB.Raw(`SELECT checkpoint FROM pipelines WHERE id = ?::uuid AND deleted_at IS NULL`, pid).Scan(&chk).Error
	var cp interface{}
	if len(chk) > 0 {
		_ = json.Unmarshal(chk, &cp)
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"checkpoint": cp})
}

func (s *State) ProcessEtlJobsStub(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{"ok": true})
}
