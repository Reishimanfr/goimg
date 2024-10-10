package main

import (
	"bash06/goimg/src/logger"
	"bash06/goimg/src/worker"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	ovhEndpoint      = flag.String("ovh-endpoint", "", "Endpoint to the OVH container")
	ovhAccessToken   = flag.String("ovh-access-token", "", "Your OVH access token")
	ovhSecretKey     = flag.String("ovh-secret-key", "", "Your OVH secret key")
	ovhRegion        = flag.String("ovh-region", "", "OVH container endpoint (like 'waw' for example)")
	ovhContainerName = flag.String("ovh-container-name", "", "Name of the container to store files in")
)

func main() {
	flag.Parse()

	// TODO: add flag for prod
	logger, err := logger.New(nil, false)
	if err != nil {
		panic(fmt.Errorf("Failed to initialize zap logger: %v", err))
	}

	if *ovhAccessToken == "" {
		logger.Fatal("No OVH access token provided")
	}

	if *ovhSecretKey == "" {
		logger.Fatal("No OVH secret key provided")
	}

	if *ovhContainerName == "" {
		logger.Fatal("No OVH container name provided")
	}

	if *ovhEndpoint == "" {
		logger.Fatal("No OVH endpoint provided")
	}

	if *ovhRegion == "" {
		logger.Fatal("No OVH region provided")
	}

	logger.Info("Establishing connection with OVH")

	_, err = worker.New(*ovhAccessToken, *ovhSecretKey, *ovhRegion, *ovhEndpoint)
	if err != nil {
		logger.Fatal("Failed to initialize OVH worker", zap.Error(err))
	}

	r := gin.New()
	r.Use(gin.Logger())

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

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	<-s

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Gracefully shutting down server")
}

// go run . -ovh-access-token="a2aa10c1a11342379a3f1255958d4e86" -ovh-secret-key="25aa998b68174e2e9bd319f470a8a60a" -ovh-endpoint="https://s3.waw.io.cloud.ovh.net/" -ovh-region="waw" -ovh-container-name="goimg"
