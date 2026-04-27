package uiassets

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist dist/*
var embeddedAssets embed.FS

func Register(engine *gin.Engine) {
	distFS, err := fs.Sub(embeddedAssets, "dist")
	if err != nil {
		return
	}

	fileServer := http.FileServer(http.FS(distFS))
	engine.GET("/", serve(distFS, fileServer))
	engine.NoRoute(serve(distFS, fileServer))
}

func serve(distFS fs.FS, fileServer http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet || strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/ws/") || strings.HasPrefix(c.Request.URL.Path, "/envs/") || strings.HasPrefix(c.Request.URL.Path, "/contest-envs/") || strings.HasPrefix(c.Request.URL.Path, "/battle-workspace/") {
			c.Status(http.StatusNotFound)
			return
		}

		if hasEmbeddedFile(distFS, c.Request.URL.Path) {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

func hasEmbeddedFile(distFS fs.FS, requestPath string) bool {
	cleanPath := strings.TrimPrefix(path.Clean("/"+requestPath), "/")
	if cleanPath == "" || cleanPath == "." {
		cleanPath = "index.html"
	}

	info, err := fs.Stat(distFS, cleanPath)
	return err == nil && !info.IsDir()
}
