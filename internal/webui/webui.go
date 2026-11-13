// Package webui embeds the prebuilt frontend static export so the API and the
// web UI ship as a single binary.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var embedded embed.FS

const notBuiltHTML = `<!doctype html>
<html lang="en">
  <head><meta charset="utf-8" /><title>Frontend not built</title></head>
  <body>
    <main>
      <h1>Frontend assets are not embedded</h1>
      <p>Run <code>make web</code> before building the binary, or run the frontend separately with <code>pnpm dev</code>.</p>
    </main>
  </body>
</html>`

// Handler serves the embedded frontend build. If the build output is missing
// (only the placeholder is embedded), it returns a short explanation instead.
func Handler() http.Handler {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		return notBuiltHandler()
	}

	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return notBuiltHandler()
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Next.js emits content-hashed assets under /_next/static.
		if strings.HasPrefix(r.URL.Path, "/_next/static/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		fileServer.ServeHTTP(w, r)
	})
}

func notBuiltHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(notBuiltHTML))
	})
}
