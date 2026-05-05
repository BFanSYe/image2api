package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kleinai/backend/internal/model"
)

func TestProviderCooldownGrokForbiddenIsTransient(t *testing.T) {
	err := errors.New(`grok upload HTTP 403: <!DOCTYPE html><html><head><title>Just a moment...</title></head></html>`)
	if got := providerCooldown(err); got != 0 {
		t.Fatalf("expected transient cooldown 0, got %s", got)
	}
}

func TestProviderCooldownRetryable429StillCooldowns(t *testing.T) {
	err := errors.New(`provider call: grok video HTTP 429: {"error":{"code":8,"message":"Too many requests"}}`)
	got := providerCooldown(err)
	if got < 30*time.Minute {
		t.Fatalf("expected 429 cooldown >= 30m, got %s", got)
	}
}

func TestValidateGenerationCreateRequestAllowsGPTImage2Resolutions(t *testing.T) {
	for _, tier := range []string{"1K", "2K", "4K", "1", "2", "4"} {
		req := CreateRequest{
			Kind:      "image",
			Provider:  "gpt",
			ModelCode: "gpt-image-2",
			Params:    map[string]any{"resolution": tier},
		}
		if err := validateGenerationCreateRequest(req); err != nil {
			t.Fatalf("expected %s to be allowed, got %v", tier, err)
		}
	}
}

func TestValidateGenerationCreateRequestAllowsGPTImage2Sizes(t *testing.T) {
	for _, size := range []string{"1344x768", "1664x928", "3312x1872", "3808x1632"} {
		req := CreateRequest{
			Kind:      "image",
			Provider:  "gpt",
			ModelCode: "gpt-image-2",
			Params:    map[string]any{"size": size},
		}
		if err := validateGenerationCreateRequest(req); err != nil {
			t.Fatalf("expected %s to be allowed, got %v", size, err)
		}
	}
}

func TestImageUpscaleTargetUsesMeta(t *testing.T) {
	meta := `{"upscale_size":"1664x928"}`
	gr := &model.GenerationResult{Meta: &meta}
	pt := imageUpscaleTarget(map[string]any{}, gr)
	if pt.X != 1664 || pt.Y != 928 {
		t.Fatalf("expected 1664x928 from meta, got %dx%d", pt.X, pt.Y)
	}
}

func TestCacheDataURLAssetUpscalesImage(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("KLEIN_STORAGE_ROOT", tmp)

	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	s := &GenerationService{}
	url, w, h, ok := s.cacheDataURLAsset(context.Background(), "local", dataURL, "task01", 0, false, image.Pt(4, 2))
	if !ok {
		t.Fatal("expected cacheDataURLAsset to succeed")
	}
	if w != 4 || h != 2 {
		t.Fatalf("expected upscaled 4x2, got %dx%d", w, h)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	if _, err := os.Stat(filepath.Join(tmp, strings.TrimPrefix(url, "/api/v1/gen/cached/"))); err != nil {
		t.Fatalf("expected cached file to exist: %v", err)
	}
}

func TestCacheDataURLAssetKeepsJPEGExtension(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("KLEIN_STORAGE_ROOT", tmp)

	dataURL := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString([]byte("not-a-real-jpeg"))
	s := &GenerationService{}
	url, _, _, ok := s.cacheDataURLAsset(context.Background(), "local", dataURL, "task02", 0, false, image.Point{})
	if !ok {
		t.Fatal("expected cacheDataURLAsset to succeed")
	}
	if !strings.HasSuffix(url, ".jpg") {
		t.Fatalf("expected jpeg cache URL to end with .jpg, got %s", url)
	}
}

func TestValidateGenerationCreateRequestRejectsUnsupportedGPTImage2Size(t *testing.T) {
	req := CreateRequest{
		Kind:      "image",
		Provider:  "gpt",
		ModelCode: "gpt-image-2",
		Params:    map[string]any{"size": "4096x4096"},
	}
	if err := validateGenerationCreateRequest(req); err == nil {
		t.Fatal("expected unsupported custom size to be rejected")
	}
}
