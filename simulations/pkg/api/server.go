package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/chainspace/simulations/pkg/engine"
	"github.com/gin-gonic/gin"
)

// Server API服务器
type Server struct {
	engine     *engine.Engine
	router     *gin.Engine
	httpServer *http.Server
	wsHub      *WebSocketHub
	config     ServerConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port       int    `json:"port"`
	Host       string `json:"host"`
	EnableCORS bool   `json:"enable_cors"`
	Debug      bool   `json:"debug"`
}

// NewServer 创建服务器
func NewServer(eng *engine.Engine, config ServerConfig) *Server {
	if config.Port == 0 {
		config.Port = 8081
	}
	if config.Host == "" {
		config.Host = "0.0.0.0"
	}

	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	server := &Server{
		engine: eng,
		router: router,
		config: config,
		wsHub:  NewWebSocketHub(eng),
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware 设置中间件
func (s *Server) setupMiddleware() {
	// Logger中间件
	s.router.Use(gin.Logger())

	// CORS中间件
	if s.config.EnableCORS {
		s.router.Use(corsMiddleware())
	}
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		// 模拟器元信息
		api.GET("/meta", s.handleGetMeta)
		api.GET("/simulators", s.handleListSimulators)

		// 模拟器生命周期
		api.POST("/simulator/init", s.handleInit)
		api.POST("/simulator/start", s.handleStart)
		api.POST("/simulator/stop", s.handleStop)
		api.POST("/simulator/pause", s.handlePause)
		api.POST("/simulator/resume", s.handleResume)
		api.POST("/simulator/step", s.handleStep)
		api.POST("/simulator/reset", s.handleReset)
		api.POST("/simulator/switch", s.handleSwitch)

		// 状态和事件
		api.GET("/simulator/state", s.handleGetState)
		api.GET("/simulator/events", s.handleGetEvents)

		// 参数
		api.GET("/simulator/params", s.handleGetParams)
		api.PUT("/simulator/params/:key", s.handleSetParam)
		api.PUT("/simulator/speed", s.handleSetSpeed)
		api.POST("/simulator/action", s.handleExecuteAction)

		// 故障注入
		api.POST("/simulator/fault", s.handleInjectFault)
		api.DELETE("/simulator/fault/:id", s.handleRemoveFault)
		api.DELETE("/simulator/faults", s.handleClearFaults)

		// 攻击注入
		api.POST("/simulator/attack", s.handleInjectAttack)
		api.DELETE("/simulator/attack/:id", s.handleRemoveAttack)
		api.DELETE("/simulator/attacks", s.handleClearAttacks)

		// 快照
		api.POST("/simulator/snapshot", s.handleSaveSnapshot)
		api.GET("/simulator/snapshots", s.handleListSnapshots)
		api.POST("/simulator/snapshot/load", s.handleLoadSnapshot)
		api.DELETE("/simulator/snapshot/:id", s.handleDeleteSnapshot)
	}

	// WebSocket
	s.router.GET("/ws/simulator", s.wsHub.HandleWebSocket)

	// 健康检查
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 启动WebSocket Hub
	go s.wsHub.Run()

	// 订阅引擎事件并推送到WebSocket
	go s.subscribeEngineEvents()

	fmt.Printf("Server starting on %s\n", addr)
	return s.httpServer.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	s.wsHub.Stop()
	return s.httpServer.Shutdown(ctx)
}

// subscribeEngineEvents 订阅引擎事件
func (s *Server) subscribeEngineEvents() {
	eventBus := s.engine.GetEventBus()
	ch := eventBus.SubscribeAll()

	for event := range ch {
		s.wsHub.BroadcastEvent(event)
	}
}

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// success 成功响应
func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// fail 失败响应
func fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// errorResponse 错误响应
func errorResponse(c *gin.Context, httpCode int, code int, message string) {
	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
