package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"image/png"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

var (
	mu          sync.Mutex
	cachedBytes []byte
	cachedName  string
	cachedAt    time.Time
)

func getOrRenderImage(cfg Config) (pngBytes []byte, name string, err error) {
	mu.Lock()
	defer mu.Unlock()

	if cachedBytes != nil && time.Since(cachedAt) < time.Duration(cfg.RefreshRate)*time.Second {
		return cachedBytes, cachedName, nil
	}

	pngBytes, name, err = renderImage(cfg)
	if err != nil {
		return nil, "", err
	}
	cachedBytes = pngBytes
	cachedName = name
	cachedAt = time.Now()
	return pngBytes, name, nil
}

func renderImage(cfg Config) (pngBytes []byte, name string, err error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	var rawPNG []byte
	if err = chromedp.Run(ctx,
		chromedp.EmulateViewport(800, 480),
		chromedp.Navigate(cfg.RenderURL),
		chromedp.WaitReady("body"),
		chromedp.CaptureScreenshot(&rawPNG),
	); err != nil {
		return nil, "", fmt.Errorf("chromedp: %w", err)
	}

	src, err := png.Decode(bytes.NewReader(rawPNG))
	if err != nil {
		return nil, "", fmt.Errorf("decode screenshot: %w", err)
	}

	dithered := ditherToSpectra6(src)

	var buf bytes.Buffer
	if err = png.Encode(&buf, dithered); err != nil {
		return nil, "", fmt.Errorf("encode dithered image: %w", err)
	}

	hash := md5.Sum(rawPNG)
	name = fmt.Sprintf("screen_%s_%x.png", time.Now().Format("20060102_150405"), hash[:4])
	return buf.Bytes(), name, nil
}
