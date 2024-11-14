package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"reishimanfr/goimg/database"
	"reishimanfr/goimg/flags"
	"reishimanfr/goimg/routes"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func watchFiles(db *gorm.DB, logger *zap.Logger) {
	path := filepath.Join(flags.BasePth, "files")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	if err := watcher.Add(path); err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if (event.Op&fsnotify.Remove == fsnotify.Remove) || (event.Op&fsnotify.Rename == fsnotify.Rename) {
				if _, err := os.Stat(event.Name); os.IsNotExist(err) {
					name := filepath.Base(event.Name)

					if err := db.Delete(&database.FileRecord{}, "filename = ?", name).Error; err != nil {
						logger.Error("Error while auto-deleting record", zap.String("Filename", event.Name), zap.Error(err))
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		}
	}
}

func main() {
	flag.Parse()

	config := zap.NewDevelopmentConfig()
	logger, err := config.Build()
	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %v", err))
	}

	fileDirPath := filepath.Join(flags.BasePth, "files")

	// Create file directory if it doesn't exist
	if _, err := os.Stat(fileDirPath); errors.Is(err, fs.ErrNotExist) {
		if err := os.Mkdir(fileDirPath, 0755); err != nil {
			logger.Fatal("Failed to create file directory", zap.String("Path", fileDirPath), zap.Error(err))
		}

		logger.Debug("Created file directory", zap.String("Path", fileDirPath))
	}

	db, err := database.Init()
	if err != nil {
		logger.Fatal("Error while initializing database", zap.Error(err))
	}

	r := gin.Default()

	server := &http.Server{
		Addr:    ":" + *flags.Port,
		Handler: r,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12, // Enforce TLS 1.2 or higher
		},
	}

	if *flags.Secure {
		server.Addr = ":443"
	}

	if !*flags.Dev {
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = io.Discard
	}

	h := routes.Handler{
		Router: r,
		Logger: logger,
		Db:     db,
	}

	h.Init()

	go func() {
		if *flags.Secure {
			if err := server.ListenAndServeTLS(*flags.SslCertPath, *flags.SslKeyPath); err != nil && err != http.ErrServerClosed {
				log.Fatal("Failed to start server", zap.Error(err))
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal("Failed to start server", zap.Error(err))
			}
		}
	}()

	logger.Info("Server started on " + server.Addr)

	watchFiles(db, logger)

	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt)
	<-q

	logger.Info("Gracefully shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}
}
