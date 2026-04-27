package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/chainspace/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type workspaceRuntimeTarget struct {
	EnvID   string
	PodName string
	Status  string
}

func proxyWorkspaceToPod(
	c *gin.Context,
	workspaceService *service.WorkspaceAccessService,
	target *workspaceRuntimeTarget,
	port int,
	notReadyMessage string,
	connectErrorMessage string,
) {
	if target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	if target.PodName == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": notReadyMessage})
		return
	}

	proxy, err := workspaceService.BuildWorkspaceProxy(
		&service.WorkspaceRuntimeTarget{
			EnvID:   target.EnvID,
			PodName: target.PodName,
			Status:  target.Status,
		},
		port,
		emptyWorkspacePath(c.Param("path")),
		c.Request.URL.RawQuery,
		c.Request.Header,
	)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": connectErrorMessage})
		return
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(connectErrorMessage))
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func handleWorkspaceFiles(
	c *gin.Context,
	workspaceService *service.WorkspaceAccessService,
	target *workspaceRuntimeTarget,
	notReadyMessage string,
) {
	if target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	if target.PodName == "" || (target.Status != "" && target.Status != "running") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": notReadyMessage})
		return
	}

	serviceTarget := &service.WorkspaceRuntimeTarget{
		EnvID:   target.EnvID,
		PodName: target.PodName,
		Status:  target.Status,
	}
	routePath := strings.TrimPrefix(c.Param("path"), "/")

	switch {
	case c.Request.Method == http.MethodGet && routePath == "api/files":
		targetPath, ok := normalizeWorkspacePath(c.Query("path"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace path"})
			return
		}
		files, err := workspaceService.ListWorkspaceFiles(c.Request.Context(), serviceTarget, targetPath)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to list workspace files"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"files": files})

	case c.Request.Method == http.MethodPost && routePath == "api/files/mkdir":
		var body struct {
			Path string `json:"path"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid create directory request"})
			return
		}
		targetPath, ok := normalizeWorkspacePath(body.Path)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace path"})
			return
		}
		if err := workspaceService.CreateWorkspaceDirectory(c.Request.Context(), serviceTarget, targetPath); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to create directory"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})

	case c.Request.Method == http.MethodDelete && routePath == "api/files":
		targetPath, ok := normalizeWorkspacePath(c.Query("path"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace path"})
			return
		}
		if targetPath == "/workspace" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete workspace root"})
			return
		}
		if err := workspaceService.DeleteWorkspacePath(c.Request.Context(), serviceTarget, targetPath); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to delete path"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})

	case c.Request.Method == http.MethodGet && routePath == "api/files/download":
		targetPath, ok := normalizeWorkspacePath(c.Query("path"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace path"})
			return
		}
		data, fileName, err := workspaceService.DownloadWorkspaceFile(c.Request.Context(), serviceTarget, targetPath)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read workspace file"})
			return
		}
		c.Header("Content-Disposition", `attachment; filename="`+fileName+`"`)
		c.Data(http.StatusOK, "application/octet-stream", data)

	case c.Request.Method == http.MethodPost && routePath == "api/files/upload":
		targetDir, ok := normalizeWorkspacePath(c.Query("path"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace path"})
			return
		}
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing upload file"})
			return
		}
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open upload file"})
			return
		}
		defer src.Close()

		data, err := io.ReadAll(src)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read upload file"})
			return
		}

		targetFile := strings.TrimRight(targetDir, "/") + "/" + file.Filename
		if err := workspaceService.UploadWorkspaceFile(c.Request.Context(), serviceTarget, targetFile, data); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to upload workspace file"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})

	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "unsupported file endpoint"})
	}
}

func handleWorkspaceLogs(
	c *gin.Context,
	workspaceService *service.WorkspaceAccessService,
	target *workspaceRuntimeTarget,
	notReadyMessage string,
	containerFetchErrorMessage string,
) {
	if target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	if target.PodName == "" || (target.Status != "" && target.Status != "running") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": notReadyMessage})
		return
	}
	if strings.TrimPrefix(c.Param("path"), "/") != "api/logs" {
		c.JSON(http.StatusNotFound, gin.H{"error": "unsupported logs endpoint"})
		return
	}

	source := strings.TrimSpace(c.Query("source"))
	levels := strings.Split(c.Query("levels"), ",")
	logs, err := workspaceService.GetWorkspaceLogs(
		c.Request.Context(),
		&service.WorkspaceRuntimeTarget{
			EnvID:   target.EnvID,
			PodName: target.PodName,
			Status:  target.Status,
		},
		source,
		levels,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": containerFetchErrorMessage})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func emptyWorkspacePath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}
