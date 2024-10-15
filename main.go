package main

import (
	"bash06/goimg/src/database"
	"bash06/goimg/src/flags"
	"bash06/goimg/src/logger"
	"bash06/goimg/src/routes"
	"bash06/goimg/src/worker"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	flag.Parse()

	// TODO: add flag for prod
	logger, err := logger.New(nil, false)
	if err != nil {
		panic(fmt.Errorf("Failed to initialize zap logger: %v", err))
	}

	if *flags.HMACSecretKey == "" {
		logger.Fatal("No HMAC secret key provided. If you want the program to generate one for you set the -generate-hmac-key flag to true")
	}

	if *flags.EnableOvhServer {
		logger.Debug("OVH server enabled. Checking for required variables now...")

		if *flags.OvhAccessToken == "" {
			logger.Fatal("No OVH access token provided")
		}

		if *flags.OvhSecretKey == "" {
			logger.Fatal("No OVH secret key provided")
		}

		if *flags.OvhContainerName == "" {
			logger.Fatal("No OVH container name provided")
		}

		if *flags.OvhEndpoint == "" {
			logger.Fatal("No OVH endpoint provided")
		}

		if *flags.OvhRegion == "" {
			logger.Fatal("No OVH region provided")
		}
	}

	// TODO: add an option to save files locally
	w, err := worker.New(*flags.OvhAccessToken, *flags.OvhSecretKey, *flags.OvhRegion, *flags.OvhEndpoint)
	if err != nil {
		logger.Fatal("Failed to initialize OVH worker", zap.Error(err))
	}

	db, err := database.New()
	if err != nil {
		logger.Fatal("Failed to initialize sqlite database", zap.Error(err))
		return
	}

	r := gin.New()
	r.Use(gin.Logger())

	routes.InitHandler(r, db, w, logger)

	server := &http.Server{
		Addr:    ":8080", // TODO: make this a flag
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to launch http server", zap.Error(err))
		}
	}()

	logger.Info("Server up and running")

	// Function to cleanup old temporary images (sent by unregistered users or when someone checked the "expire after X" box)
	go startPeriodicCleanup(db, logger, time.Hour*1)

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	<-s

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Gracefully shutting down server")
}

func startPeriodicCleanup(db *gorm.DB, logger *zap.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	for {
		<-ticker.C
		now := time.Now().Unix()

		err := db.Where("expires_at < ?", now).Delete(&database.FileRecord{}).Error
		if err != nil {
			logger.Error("Error while deleting expired files", zap.Error(err))
		} else {
			logger.Debug("Periodic expired file cleanup finished")
		}
	}
}
