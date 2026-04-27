package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	backendapp "github.com/chainspace/backend/internal/app"
	"github.com/chainspace/backend/internal/config"
	"github.com/chainspace/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// main 负责加载配置、启动应用并处理优雅关闭。
func main() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting ChainSpace Backend Server...")

	app, err := backendapp.Build(cfg)
	if err != nil {
		logger.Fatal("Failed to build application", zap.Error(err))
	}

	app.InitPlatformAdmin(context.Background())
	app.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
