package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *EnvProxyHandler) getEnvForWorkspace(c *gin.Context) *workspaceRuntimeTarget {
	envID := c.Param("env_id")
	target, err := h.workspaceService.ResolveExperimentTarget(c.Request.Context(), envID)
	if err != nil || target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "experiment environment not found"})
		return nil
	}

	return &workspaceRuntimeTarget{
		EnvID:   target.EnvID,
		PodName: target.PodName,
		Status:  target.Status,
	}
}

func (h *EnvProxyHandler) getInstanceTarget(c *gin.Context) *workspaceRuntimeTarget {
	envID := c.Param("env_id")
	instanceKey := c.Param("instance_key")
	target, err := h.workspaceService.ResolveExperimentInstanceTarget(c.Request.Context(), envID, instanceKey)
	if err != nil || target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "experiment runtime instance not found"})
		return nil
	}

	return &workspaceRuntimeTarget{
		EnvID:   target.EnvID,
		PodName: target.PodName,
		Status:  target.Status,
	}
}

func (h *ContestEnvProxyHandler) getContestEnvForWorkspace(c *gin.Context) *workspaceRuntimeTarget {
	envID := c.Param("env_id")
	target, err := h.workspaceService.ResolveContestTarget(c.Request.Context(), envID)
	if err != nil || target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "challenge environment not found"})
		return nil
	}

	return &workspaceRuntimeTarget{
		EnvID:   target.EnvID,
		PodName: target.PodName,
		Status:  target.Status,
	}
}

func normalizeWorkspacePath(input string) (string, bool) {
	value := strings.TrimSpace(input)
	if value == "" {
		value = "/workspace"
	}

	if !strings.HasPrefix(value, "/") {
		value = "/workspace/" + strings.TrimPrefix(value, "./")
	}

	if !strings.HasPrefix(value, "/workspace") {
		return "", false
	}

	return strings.TrimRight(value, "/"), true
}
