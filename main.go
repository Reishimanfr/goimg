package main

import (
	"bash06/goimg/src/api"
	"bash06/goimg/src/auth"
	"bash06/goimg/src/database"
	"bash06/goimg/src/flags"
	"bash06/goimg/src/logger"
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	flag.Parse()

	// TODO: add flag for prod
	logger, err := logger.New(nil, false)
	if err != nil {
		panic(fmt.Errorf("Failed to initialize zap logger: %v", err))
	}

	if *flags.FileLocation == "aws" {
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

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting auth server...")

		db, err := auth.NewDb()
		if err != nil {
			logger.Fatal("Failed to initialize API SQlite database", zap.Error(err))
			return
		}

		r := gin.New()
		r.Use(gin.Logger())

		auth.InitHandler(r, db, logger)

		server := &http.Server{
			Addr:    ":8080",
			Handler: r,
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to launch auth server", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting API server...")

		db, err := database.NewAPIDb()
		if err != nil {
			logger.Fatal("Failed to initialize API SQlite database", zap.Error(err))
			return
		}

		r := gin.New()
		r.Use(gin.Logger())

		api.InitHandler(r, db, nil, logger)

		server := &http.Server{
			Addr:    ":8081",
			Handler: r,
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to launch API server", zap.Error(err))
		}
	}()

	logger.Info("Gracefully shutting down server")
}
