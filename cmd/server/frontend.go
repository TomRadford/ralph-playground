package main

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v5"
)

//go:embed all:web/dist
var frontendAssets embed.FS

func registerFrontendRoutes(e *echo.Echo) {
	distFS, err := fs.Sub(frontendAssets, "web/dist")
	if err != nil {
		panic("load embedded frontend assets: " + err.Error())
	}

	e.GET("/*", func(c *echo.Context) error {
		requested := normalizeRequestedPath(c.Request().URL.Path)
		if fileExists(distFS, requested) {
			http.ServeFileFS(c.Response(), c.Request(), distFS, requested)
			return nil
		}

		fallback := "200.html"
		if !fileExists(distFS, fallback) {
			fallback = "index.html"
		}
		if !fileExists(distFS, fallback) {
			return c.NoContent(http.StatusNotFound)
		}

		http.ServeFileFS(c.Response(), c.Request(), distFS, fallback)
		return nil
	})
}

func normalizeRequestedPath(rawPath string) string {
	cleanPath := path.Clean("/" + strings.TrimSpace(rawPath))
	trimmed := strings.TrimPrefix(cleanPath, "/")
	if trimmed == "" || trimmed == "." {
		return "index.html"
	}
	return trimmed
}

func fileExists(filesystem fs.FS, filePath string) bool {
	file, err := filesystem.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false
	}
	return !info.IsDir()
}
