package main

import (
	"flag"
	"log"
)

type Config struct {
	RenderURL   string
	AccessToken string
	RefreshRate int
	ListenAddr  string
	BaseURL     string
}

func parseConfig() Config {
	var cfg Config
	flag.StringVar(&cfg.RenderURL, "url", "", "URL to render (required)")
	flag.StringVar(&cfg.AccessToken, "token", "", "Expected access-token header value (omit to disable authentication)")
	flag.IntVar(&cfg.RefreshRate, "refresh", 1800, "Cache TTL and refresh_rate in seconds")
	flag.StringVar(&cfg.ListenAddr, "addr", ":8080", "HTTP listen address")
	flag.StringVar(&cfg.BaseURL, "base-url", "http://localhost:8080", "Base URL for image links (include port if non-standard)")
	flag.Parse()

	if cfg.RenderURL == "" {
		log.Fatal("-url is required")
	}
	return cfg
}
