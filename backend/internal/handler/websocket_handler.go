package handler

import (
	"fmt"
	"net/http"

	"github.com/chainspace/backend/internal/middleware"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/pkg/websocket"
	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应该严格检查Origin
	},
}

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	hub *websocket.Hub
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(hub *websocket.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// Connect 建立WebSocket连接
func (h *WebSocketHandler) Connect(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	schoolID, _ := middleware.GetSchoolID(c)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	client := h.hub.RegisterClient(conn, userID, schoolID)

	// 自动加入用户房间
	h.hub.JoinRoom(client, userRoom(userID))
	if schoolID > 0 {
		h.hub.JoinRoom(client, schoolRoom(schoolID))
	}

	go client.WritePump()
	go client.ReadPump()
}

// JoinEnvRoom 加入实验环境房间
func (h *WebSocketHandler) JoinEnvRoom(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	envID := c.Param("env_id")

	client := h.getClientByUserID(userID)
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not connected"})
		return
	}

	h.hub.JoinRoom(client, envRoom(envID))
	c.JSON(http.StatusOK, gin.H{"message": "joined"})
}

// LeaveEnvRoom 离开实验环境房间
func (h *WebSocketHandler) LeaveEnvRoom(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	envID := c.Param("env_id")

	client := h.getClientByUserID(userID)
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not connected"})
		return
	}

	h.hub.LeaveRoom(client, envRoom(envID))
	c.JSON(http.StatusOK, gin.H{"message": "left"})
}

func (h *WebSocketHandler) JoinSessionRoom(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	sessionKey := c.Param("session_key")

	client := h.getClientByUserID(userID)
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not connected"})
		return
	}

	h.hub.JoinRoom(client, sessionRoom(sessionKey))
	c.JSON(http.StatusOK, gin.H{"message": "joined"})
}

func (h *WebSocketHandler) LeaveSessionRoom(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	sessionKey := c.Param("session_key")

	client := h.getClientByUserID(userID)
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not connected"})
		return
	}

	h.hub.LeaveRoom(client, sessionRoom(sessionKey))
	c.JSON(http.StatusOK, gin.H{"message": "left"})
}

// JoinContestRoom 加入竞赛房间
func (h *WebSocketHandler) JoinContestRoom(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	contestID := c.Param("contest_id")

	client := h.getClientByUserID(userID)
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not connected"})
		return
	}

	h.hub.JoinRoom(client, contestRoom(contestID))
	c.JSON(http.StatusOK, gin.H{"message": "joined"})
}

// LeaveContestRoom 离开竞赛房间
func (h *WebSocketHandler) LeaveContestRoom(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	contestID := c.Param("contest_id")

	client := h.getClientByUserID(userID)
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not connected"})
		return
	}

	h.hub.LeaveRoom(client, contestRoom(contestID))
	c.JSON(http.StatusOK, gin.H{"message": "left"})
}

// GetOnlineStats 获取在线统计
func (h *WebSocketHandler) GetOnlineStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"online_count": h.hub.GetOnlineCount(),
	})
}

// 内部方法

func (h *WebSocketHandler) getClientByUserID(userID uint) *websocket.Client {
	return h.hub.GetClientByUserID(userID)
}

func userRoom(userID uint) string {
	return fmt.Sprintf("user:%d", userID)
}

func schoolRoom(schoolID uint) string {
	return fmt.Sprintf("school:%d", schoolID)
}

func envRoom(envID string) string {
	return "env:" + envID
}

func contestRoom(contestID string) string {
	return "contest:" + contestID
}

func sessionRoom(sessionKey string) string {
	return "session:" + sessionKey
}
