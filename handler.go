package main

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type displayResponse struct {
	Status         int     `json:"status"`
	ImageURL       string  `json:"image_url"`
	Filename       string  `json:"filename"`
	UpdateFirmware bool    `json:"update_firmware"`
	FirmwareURL    *string `json:"firmware_url"`
	RefreshRate    int     `json:"refresh_rate"`
	ResetFirmware  bool    `json:"reset_firmware"`
}

type setupResponse struct {
	Status     int     `json:"status"`
	APIKey     *string `json:"api_key"`
	FriendlyID *string `json:"friendly_id"`
	ImageURL   *string `json:"image_url"`
	Filename   *string `json:"filename"`
	Message    string  `json:"message"`
}

func setupHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mac := r.Header.Get("ID")

		friendlyID := macToFriendlyID(mac)
		apiKey := cfg.AccessToken

		var resp setupResponse
		if mac == "" {
			resp = setupResponse{
				Status:  404,
				Message: "MAC Address not registered",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(resp)
			return
		}

		logOut.Printf("setup: device MAC=%s friendly_id=%s", mac, friendlyID)

		resp = setupResponse{
			Status:     200,
			APIKey:     &apiKey,
			FriendlyID: &friendlyID,
			ImageURL:   nil,
			Message:    "Welcome to gTRMNL",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// macToFriendlyID derives a short uppercase ID from the last 3 MAC octets.
// e.g. "24:fb:e3:44:95:a1" → "4495A1"
func macToFriendlyID(mac string) string {
	parts := strings.Split(mac, ":")
	if len(parts) < 3 {
		return strings.ToUpper(strings.ReplaceAll(mac, ":", ""))
	}
	last3 := parts[len(parts)-3:]
	return strings.ToUpper(strings.Join(last3, ""))
}

func displayHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.AccessToken != "" {
			token := r.Header.Get("access-token")
			if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.AccessToken)) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		_, name, err := getOrRenderImage(cfg)
		if err != nil {
			logErr.Printf("render error: %v", err)
			http.Error(w, "render failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resp := displayResponse{
			Status:         0,
			ImageURL:       cfg.BaseURL + "/image/" + name,
			Filename:       name,
			UpdateFirmware: false,
			FirmwareURL:    nil,
			RefreshRate:    cfg.RefreshRate,
			ResetFirmware:  false,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func logHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.Header.Get("ID")

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			logErr.Printf("log decode error (device %s): %v", deviceID, err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		pretty, _ := json.MarshalIndent(body, "", "  ")
		logOut.Printf("device log (ID=%s):\n%s", deviceID, pretty)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":200}`)
	}
}

func previewHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, name, err := getOrRenderImage(cfg)
		if err != nil {
			logErr.Printf("preview render error: %v", err)
			http.Error(w, "render failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		http.ServeContent(w, r, name, cachedAt, bytes.NewReader(data))
	}
}

func imageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		data := cachedBytes
		name := cachedName
		at := cachedAt
		mu.Unlock()

		if data == nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		http.ServeContent(w, r, name, at, bytes.NewReader(data))
	}
}
