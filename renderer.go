package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"image/png"
	"sync"
	"time"

	"github.com/chromedp/cdproto/emulation"
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
	return
}

func renderImage(cfg Config) (pngBytes []byte, name string, err error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	if cfg.ChromePath != "" {
		opts = append(opts, chromedp.ExecPath(cfg.ChromePath))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	forceLight := chromedp.ActionFunc(func(ctx context.Context) error {
		return emulation.SetEmulatedMedia().
			WithMedia("screen").
			WithFeatures([]*emulation.MediaFeature{
				{Name: "prefers-color-scheme", Value: "light"},
			}).Do(ctx)
	})

	var rawPNG []byte
	if err = chromedp.Run(ctx,
		chromedp.EmulateViewport(800, 480),
		forceLight,
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
		return nil, "", fmt.Errorf("encode PNG: %w", err)
	}

	hash := md5.Sum(rawPNG)
	name = fmt.Sprintf("screen_%s_%x.png", time.Now().Format("20060102_150405"), hash[:4])
	return buf.Bytes(), name, nil
}
