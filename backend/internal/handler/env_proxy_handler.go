package handler

import (
	"net/http"

	"github.com/chainspace/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type EnvProxyHandler struct {
	workspaceService *service.WorkspaceAccessService
}

func NewEnvProxyHandler(workspaceService *service.WorkspaceAccessService) *EnvProxyHandler {
	return &EnvProxyHandler{workspaceService: workspaceService}
}

func (h *EnvProxyHandler) ProxyIDE(c *gin.Context) {
	h.proxyToTool(c, "ide", 8443)
}

func (h *EnvProxyHandler) ProxyTerminal(c *gin.Context) {
	h.proxyToTool(c, "terminal", 7681)
}

func (h *EnvProxyHandler) ProxyRPC(c *gin.Context) {
	h.proxyToTool(c, "rpc", 8545)
}

func (h *EnvProxyHandler) ProxyAPIDebug(c *gin.Context) {
	h.proxyToTool(c, "api_debug", 6688)
}

func (h *EnvProxyHandler) ProxySim(c *gin.Context) {
	h.proxyToTool(c, "visualization", 8080)
}

func (h *EnvProxyHandler) ProxyExplorer(c *gin.Context) {
	h.proxyToTool(c, "explorer", 4000)
}

func (h *EnvProxyHandler) ProxyNetwork(c *gin.Context) {
	h.proxyToTool(c, "network", 7681)
}

func (h *EnvProxyHandler) ProxyFiles(c *gin.Context) {
	handleWorkspaceFiles(c, h.workspaceService, h.getToolTarget(c, "files"), "experiment environment is not ready")
}

func (h *EnvProxyHandler) ProxyLogs(c *gin.Context) {
	handleWorkspaceLogs(c, h.workspaceService, h.getToolTarget(c, "logs"), "experiment environment is not ready", "failed to fetch experiment container")
}

func (h *EnvProxyHandler) proxyToTool(c *gin.Context, toolKey string, _ int) {
	target, port := h.resolveToolTarget(c, toolKey)
	if target == nil {
		return
	}
	proxyWorkspaceToPod(c, h.workspaceService, target, port, "experiment environment is not ready", "experiment environment connection failed")
}

func (h *EnvProxyHandler) resolveToolTarget(c *gin.Context, toolKey string) (*workspaceRuntimeTarget, int) {
	envID := c.Param("env_id")
	target, port, err := h.workspaceService.ResolveExperimentToolTarget(c.Request.Context(), envID, toolKey)
	if err == nil && target != nil {
		return &workspaceRuntimeTarget{
			EnvID:   target.EnvID,
			PodName: target.PodName,
			Status:  target.Status,
		}, port
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "experiment tool target not found"})
	return nil, 0
}

func (h *EnvProxyHandler) getToolTarget(c *gin.Context, toolKey string) *workspaceRuntimeTarget {
	target, _ := h.resolveToolTarget(c, toolKey)
	return target
}

func (h *EnvProxyHandler) ProxyInstanceIDE(c *gin.Context) {
	h.proxyToInstanceTool(c, "ide", 8443)
}

func (h *EnvProxyHandler) ProxyInstanceTerminal(c *gin.Context) {
	h.proxyToInstanceTool(c, "terminal", 7681)
}

func (h *EnvProxyHandler) ProxyInstanceRPC(c *gin.Context) {
	h.proxyToInstanceTool(c, "rpc", 8545)
}

func (h *EnvProxyHandler) ProxyInstanceAPIDebug(c *gin.Context) {
	h.proxyToInstanceTool(c, "api_debug", 6688)
}

func (h *EnvProxyHandler) ProxyInstanceSim(c *gin.Context) {
	h.proxyToInstanceTool(c, "visualization", 8080)
}

func (h *EnvProxyHandler) ProxyInstanceExplorer(c *gin.Context) {
	h.proxyToInstanceTool(c, "explorer", 4000)
}

func (h *EnvProxyHandler) ProxyInstanceNetwork(c *gin.Context) {
	h.proxyToInstanceTool(c, "network", 7681)
}

func (h *EnvProxyHandler) ProxyInstanceFiles(c *gin.Context) {
	handleWorkspaceFiles(c, h.workspaceService, h.getInstanceTarget(c), "experiment runtime instance is not ready")
}

func (h *EnvProxyHandler) ProxyInstanceLogs(c *gin.Context) {
	handleWorkspaceLogs(c, h.workspaceService, h.getInstanceTarget(c), "experiment runtime instance is not ready", "failed to fetch experiment instance container")
}

func (h *EnvProxyHandler) proxyToInstanceTool(c *gin.Context, toolKey string, fallbackPort int) {
	target := h.getInstanceTarget(c)
	if target == nil {
		return
	}
	port := fallbackPort
	envID := c.Param("env_id")
	instanceKey := c.Param("instance_key")
	if resolvedTarget, resolvedPort, err := h.workspaceService.ResolveExperimentInstanceToolTarget(c.Request.Context(), envID, instanceKey, toolKey); err == nil && resolvedTarget != nil && resolvedPort > 0 {
		target = &workspaceRuntimeTarget{
			EnvID:   resolvedTarget.EnvID,
			PodName: resolvedTarget.PodName,
			Status:  resolvedTarget.Status,
		}
		port = resolvedPort
	}
	proxyWorkspaceToPod(c, h.workspaceService, target, port, "experiment runtime instance is not ready", "experiment runtime instance connection failed")
}
