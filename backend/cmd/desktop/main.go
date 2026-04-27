package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	backendapp "github.com/chainspace/backend/internal/app"
	"github.com/chainspace/backend/internal/config"
	"github.com/chainspace/backend/internal/desktop"
	"github.com/chainspace/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	bootstrap, err := desktop.New(desktop.Config{
		ServerURL: "http://127.0.0.1:3000",
	})
	if err != nil {
		desktop.ShowErrorDialog("ChainSpace 启动失败", fmt.Sprintf("桌面启动准备失败：%v", err))
		fmt.Printf("Failed to prepare desktop bootstrap: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := bootstrap.Prepare(ctx); err != nil {
		desktop.ShowErrorDialog("ChainSpace 需要 Docker Desktop", "未检测到可用的 Docker Desktop。\n\n请先启动 Docker Desktop，等待其完全就绪后再重新打开 ChainSpace。")
		fmt.Printf("Failed to prepare infrastructure: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load(filepath.Join(bootstrap.ProjectRoot(), "backend", "configs", "config.desktop.yaml"))
	if err != nil {
		desktop.ShowErrorDialog("ChainSpace 启动失败", fmt.Sprintf("桌面配置加载失败：%v", err))
		fmt.Printf("Failed to load desktop config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(&cfg.Log); err != nil {
		desktop.ShowErrorDialog("ChainSpace 启动失败", fmt.Sprintf("日志初始化失败：%v", err))
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	app, err := backendapp.Build(cfg)
	if err != nil {
		desktop.ShowErrorDialog("ChainSpace 启动失败", fmt.Sprintf("后端应用启动失败：%v", err))
		logger.Fatal("Failed to build application", zap.Error(err))
	}

	app.InitPlatformAdmin(context.Background())
	app.Start()

	serverCtx, serverCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer serverCancel()
	if err := bootstrap.WaitForServer(serverCtx); err != nil {
		desktop.ShowErrorDialog("ChainSpace 启动失败", fmt.Sprintf("服务未在预期时间内就绪：%v", err))
		logger.Fatal("Server did not become ready", zap.Error(err))
	}

	if err := bootstrap.OpenUI(); err != nil {
		logger.Warn("Failed to open browser automatically", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	if err := app.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}
}
