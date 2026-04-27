package handler

import (
	"net/http"

	"github.com/chainspace/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ContestEnvProxyHandler proxies contest challenge workspace traffic through the runtime service.
type ContestEnvProxyHandler struct {
	workspaceService *service.WorkspaceAccessService
}

func NewContestEnvProxyHandler(workspaceService *service.WorkspaceAccessService) *ContestEnvProxyHandler {
	return &ContestEnvProxyHandler{workspaceService: workspaceService}
}

func (h *ContestEnvProxyHandler) ProxyIDE(c *gin.Context) {
	h.proxyToContestTool(c, "ide", 8443)
}

func (h *ContestEnvProxyHandler) ProxyTerminal(c *gin.Context) {
	h.proxyToContestTool(c, "terminal", 7681)
}

func (h *ContestEnvProxyHandler) ProxyRPC(c *gin.Context) {
	h.proxyToContestTool(c, "rpc", 8545)
}

func (h *ContestEnvProxyHandler) ProxyAPIDebug(c *gin.Context) {
	h.proxyToContestTool(c, "api_debug", 6688)
}

func (h *ContestEnvProxyHandler) ProxySim(c *gin.Context) {
	h.proxyToContestTool(c, "visualization", 8080)
}

func (h *ContestEnvProxyHandler) ProxyExplorer(c *gin.Context) {
	h.proxyToContestTool(c, "explorer", 4000)
}

func (h *ContestEnvProxyHandler) ProxyNetwork(c *gin.Context) {
	h.proxyToContestTool(c, "network", 7681)
}

func (h *ContestEnvProxyHandler) ProxyService(c *gin.Context) {
	envID := c.Param("env_id")
	serviceKey := c.Param("service_key")
	target, port, err := h.workspaceService.ResolveContestServiceTarget(c.Request.Context(), envID, serviceKey)
	if err != nil || target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "challenge service not found"})
		return
	}

	proxyWorkspaceToPod(c, h.workspaceService, &workspaceRuntimeTarget{
		EnvID:   target.EnvID,
		PodName: target.PodName,
		Status:  target.Status,
	}, port, "challenge service is not ready", "challenge service connection failed")
}

func (h *ContestEnvProxyHandler) ProxyFiles(c *gin.Context) {
	handleWorkspaceFiles(c, h.workspaceService, h.getContestEnvForWorkspace(c), "challenge environment is not ready")
}

func (h *ContestEnvProxyHandler) ProxyLogs(c *gin.Context) {
	handleWorkspaceLogs(c, h.workspaceService, h.getContestEnvForWorkspace(c), "challenge environment is not ready", "failed to fetch challenge container")
}

func (h *ContestEnvProxyHandler) proxyToContestTool(c *gin.Context, toolKey string, fallbackPort int) {
	envID := c.Param("env_id")
	target, port, err := h.workspaceService.ResolveContestToolTarget(c.Request.Context(), envID, toolKey)
	if err != nil || target == nil {
		proxyWorkspaceToPod(c, h.workspaceService, h.getContestEnvForWorkspace(c), fallbackPort, "challenge environment is not ready", "challenge environment connection failed")
		return
	}

	proxyWorkspaceToPod(c, h.workspaceService, &workspaceRuntimeTarget{
		EnvID:   target.EnvID,
		PodName: target.PodName,
		Status:  target.Status,
	}, port, "challenge environment is not ready", "challenge environment connection failed")
}
