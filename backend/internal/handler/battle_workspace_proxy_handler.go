package handler

import (
	"net/http"

	"github.com/chainspace/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// BattleWorkspaceProxyHandler proxies agent-battle team workspace traffic.
// This handler is intentionally separated from experiment env proxy to avoid
// coupling battle workspace routing with experiment env repositories.
type BattleWorkspaceProxyHandler struct {
	workspaceService *service.WorkspaceAccessService
}

func NewBattleWorkspaceProxyHandler(workspaceService *service.WorkspaceAccessService) *BattleWorkspaceProxyHandler {
	return &BattleWorkspaceProxyHandler{workspaceService: workspaceService}
}

func (h *BattleWorkspaceProxyHandler) ProxyIDE(c *gin.Context) {
	proxyWorkspaceToPod(c, h.workspaceService, h.getBattleWorkspaceTarget(c), 8443, "battle workspace is not ready", "battle workspace IDE connection failed")
}

func (h *BattleWorkspaceProxyHandler) ProxyTerminal(c *gin.Context) {
	proxyWorkspaceToPod(c, h.workspaceService, h.getBattleWorkspaceTarget(c), 7681, "battle workspace is not ready", "battle workspace terminal connection failed")
}

func (h *BattleWorkspaceProxyHandler) ProxyRPC(c *gin.Context) {
	proxyWorkspaceToPod(c, h.workspaceService, h.getBattleWorkspaceTarget(c), 8545, "battle workspace is not ready", "battle workspace RPC connection failed")
}

func (h *BattleWorkspaceProxyHandler) ProxyAPIDebug(c *gin.Context) {
	proxyWorkspaceToPod(c, h.workspaceService, h.getBattleWorkspaceTarget(c), 6688, "battle workspace is not ready", "battle workspace API debug connection failed")
}

func (h *BattleWorkspaceProxyHandler) ProxyFiles(c *gin.Context) {
	handleWorkspaceFiles(c, h.workspaceService, h.getBattleWorkspaceTarget(c), "battle workspace is not ready")
}

func (h *BattleWorkspaceProxyHandler) ProxyLogs(c *gin.Context) {
	handleWorkspaceLogs(c, h.workspaceService, h.getBattleWorkspaceTarget(c), "battle workspace is not ready", "failed to fetch battle workspace container")
}

func (h *BattleWorkspaceProxyHandler) ProxyExplorer(c *gin.Context) {
	proxyWorkspaceToPod(c, h.workspaceService, h.getBattleWorkspaceTarget(c), 4000, "battle workspace is not ready", "battle workspace explorer connection failed")
}

func (h *BattleWorkspaceProxyHandler) ProxyNetwork(c *gin.Context) {
	proxyWorkspaceToPod(c, h.workspaceService, h.getBattleWorkspaceTarget(c), 7681, "battle workspace is not ready", "battle workspace network panel connection failed")
}

func (h *BattleWorkspaceProxyHandler) ProxyVisualization(c *gin.Context) {
	proxyWorkspaceToPod(c, h.workspaceService, h.getBattleWorkspaceTarget(c), 8080, "battle workspace is not ready", "battle workspace visualization connection failed")
}

func (h *BattleWorkspaceProxyHandler) getBattleWorkspaceTarget(c *gin.Context) *workspaceRuntimeTarget {
	podName := c.Param("env_id")
	if podName == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "battle workspace not found"})
		return nil
	}

	return &workspaceRuntimeTarget{
		EnvID:   podName,
		PodName: podName,
		Status:  "running",
	}
}
