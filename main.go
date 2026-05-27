package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	logOut = log.New(os.Stdout, "", log.LstdFlags)
	logErr = log.New(os.Stderr, "", log.LstdFlags)
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(sw, r)
		logOut.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond))
	})
}

func main() {
	cfg := parseConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/setup", setupHandler(cfg))
	mux.HandleFunc("GET /api/setup/", setupHandler(cfg))
	mux.HandleFunc("GET /api/display", displayHandler(cfg))
	mux.HandleFunc("GET /api/display/", displayHandler(cfg))
	mux.HandleFunc("POST /api/log", logHandler())
	mux.HandleFunc("POST /api/log/", logHandler())
	mux.HandleFunc("GET /image/", imageHandler())

	logOut.Printf("gTRMNL listening on %s", cfg.ListenAddr)
	logOut.Printf("  render URL:   %s", cfg.RenderURL)
	logOut.Printf("  refresh rate: %ds", cfg.RefreshRate)
	logOut.Printf("  base URL:     %s", cfg.BaseURL)
	if strings.Contains(cfg.BaseURL, "localhost") || strings.Contains(cfg.BaseURL, "127.0.0.1") {
		logErr.Printf("WARNING: -base-url contains %q — the device will not be able to fetch images; use the server's LAN IP instead", cfg.BaseURL)
	}

	if err := http.ListenAndServe(cfg.ListenAddr, logging(mux)); err != nil {
		logErr.Fatalf("server: %v", err)
	}
}
