// @title           Messaging Server API
// @version         1.0
// @description     A simple scheduler-based messaging service.
// @host            localhost:8080
// @BasePath        /
// @schemes         http
package main

import (
	"context"
	"log"
	_ "messaging-server/docs"
	"messaging-server/internal/db"
	"messaging-server/internal/jobs"
	"messaging-server/internal/logger"
	"messaging-server/internal/router"
	"messaging-server/internal/scheduler"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// init logger
	logger.Init()
	defer logger.Sync()

	// init cache
	err := db.InitRedis()
	if err != nil {
		logger.Sugar.Fatalf("redis connection failed %v", err)
	}
	// init postgres db
	err = db.InitPostgres()
	if err != nil {
		logger.Sugar.Fatalf("postgres connection failed %v", err)
	}

	// create scheduler with the send messages job
	cron, err := scheduler.New(jobs.SendMessages)
	if err != nil {
		log.Fatalf("could not create scheduler: %v", err)
	}

	// wire up routes defined in router.go
	r := router.SetupRouter(cron)

	// build HTTP server instead of r.Run()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// start scheduler immediately
	if status, code := cron.Start(); code != http.StatusCreated {
		logger.Sugar.Warnf("%s (%d)", status, code)
	}

	// run HTTP server in background
	go func() {
		logger.Sugar.Infof("HTTP server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Sugar.Fatalf("listen error: %v", err)
		}
	}()

	// wait for SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Sugar.Info("shutdown signal received")

	// give up to 30s for cleanup gin-gonic http server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// stop scheduler
	cron.Stop()

	// shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Sugar.Fatalf("server shutdown failed: %v", err)
	}
	logger.Sugar.Info("server exited cleanly")
}
