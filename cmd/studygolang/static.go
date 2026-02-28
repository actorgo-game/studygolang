// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package main

import (
	"net/http"
	"os"
	"strings"

	echo "github.com/labstack/echo/v4"

	. "github.com/polaris1119/config"
)

type staticRootConf struct {
	root   string
	isFile bool
}

var staticFileMap = map[string]staticRootConf{
	"/static":      {"/static", false},
	"/favicon.ico": {"/static/img/go.ico", true},
	"/sitemap":     {"/sitemap", false},
}

var filterPrefixs = make([]string, 0, 3)

func serveStatic(e *echo.Echo) {
	for prefix, rootConf := range staticFileMap {
		filterPrefixs = append(filterPrefixs, prefix)

		if rootConf.isFile {
			e.File(prefix, ROOT+rootConf.root)
		} else {
			e.Static(prefix, ROOT+rootConf.root)
		}
	}

	serveFrontend(e)
}

// serveFrontend serves the Vue 3 SPA from frontend/dist/ in production.
// In development, the Vite dev server handles this via proxy.
func serveFrontend(e *echo.Echo) {
	frontendDir := ROOT + "/frontend/dist"
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		return
	}

	filterPrefixs = append(filterPrefixs, "/assets")
	e.Static("/assets", frontendDir+"/assets")

	indexHTML, err := os.ReadFile(frontendDir + "/index.html")
	if err != nil {
		return
	}

	e.GET("/app/*", func(c echo.Context) error {
		return c.HTMLBlob(http.StatusOK, indexHTML)
	})

	// SPA fallback: serve index.html for any path not matched by API/static routes
	e.Any("/*", func(c echo.Context) error {
		path := c.Request().URL.Path
		// Skip API, WebSocket, admin API and static file requests
		if strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/ws") ||
			strings.HasPrefix(path, "/static/") ||
			strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/sitemap") ||
			strings.HasPrefix(path, "/install") ||
			strings.HasPrefix(path, "/oauth/") ||
			strings.HasPrefix(path, "/captcha/") ||
			strings.HasPrefix(path, "/image/") ||
			path == "/feed.xml" {
			return echo.ErrNotFound
		}

		// Check if a real static file exists in frontend/dist
		filePath := frontendDir + path
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			return c.File(filePath)
		}

		return c.HTMLBlob(http.StatusOK, indexHTML)
	})
}
