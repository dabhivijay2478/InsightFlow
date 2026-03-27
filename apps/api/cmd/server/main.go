package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mantrixflow/go-api/internal/config"
	"github.com/mantrixflow/go-api/internal/database"
	"github.com/mantrixflow/go-api/internal/etl"
	"github.com/mantrixflow/go-api/internal/queue"
	"github.com/mantrixflow/go-api/internal/server"
	"github.com/mantrixflow/go-api/internal/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	lev, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		lev = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lev)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	fiberCfg := fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if cfg.Environment == "production" {
		fiberCfg.Prefork = false
		fiberCfg.DisableStartupMessage = true
	}

	ctx := context.Background()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("database")
	}

	var q *queue.Service
	if qs, err := queue.NewService(ctx, cfg); err != nil {
		log.Warn().Err(err).Msg("pgmq disabled — queue worker idle")
	} else {
		q = qs
	}

	etlCl := etl.New(cfg)
	st := server.NewState(cfg, db, q, etlCl)

	app := fiber.New(fiberCfg)
	app.Use(recover.New())
	app.Use(server.RequestLogger())
	server.RegisterRoutes(st, app)

	wkr := worker.New(cfg, st, q)
	wctx, wcancel := context.WithCancel(context.Background())
	defer wcancel()
	go wkr.Run(wctx)

	addr := ":" + cfg.Port
	go func() {
		log.Info().Str("addr", addr).Msg("listening")
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("listen")
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Warn().Err(err).Msg("http shutdown")
	}
	wcancel()
	if q != nil {
		q.Close()
	}
}
