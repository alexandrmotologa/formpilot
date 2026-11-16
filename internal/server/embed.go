package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var webAssets embed.FS

// DistHandler returns an http.Handler that serves the embedded web assets.
func DistHandler() http.Handler {
	sub, err := fs.Sub(webAssets, "dist")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>formpilot</title></head><body><h1>formpilot daemon running</h1><p>Web UI is building or not embedded.</p></body></html>`))
		})
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If requesting an API endpoint, do not serve static files
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to open the file in the embedded FS
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		f, err := sub.Open(path)
		if err != nil {
			// SPA fallback: serve index.html for client-side routing
			indexFile, err := sub.Open("index.html")
			if err != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>formpilot</title></head><body><h1>formpilot API Active</h1><p>Cockpit UI is available at /api/v1/health or build web/dist.</p></body></html>`))
				return
			}
			_ = indexFile.Close()
			r.URL.Path = "/"
		} else {
			_ = f.Close()
		}

		fileServer.ServeHTTP(w, r)
	})
}
