package service

import (
	"errors"
	"testing"
	"time"
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

func TestValidateGenerationCreateRequestRejectsUnsupportedGPTImage2Resolution(t *testing.T) {
	base := CreateRequest{
		Kind:      "image",
		Provider:  "gpt",
		ModelCode: "gpt-image-2",
	}

	for _, tier := range []string{"2K", "4K", "2", "4"} {
		req := base
		req.Params = map[string]any{"resolution": tier}
		if err := validateGenerationCreateRequest(req); err == nil {
			t.Fatalf("expected %s to be rejected", tier)
		}
	}
}

func TestValidateGenerationCreateRequestAllowsGPTImage21K(t *testing.T) {
	for _, params := range []map[string]any{
		{"resolution": "1K"},
		{"size_tier": "1"},
		{"size": "1344x768"},
		{},
	} {
		req := CreateRequest{Kind: "image", Provider: "gpt", ModelCode: "gpt-image-2", Params: params}
		if err := validateGenerationCreateRequest(req); err != nil {
			t.Fatalf("expected params %#v to be allowed, got %v", params, err)
		}
	}
}

func TestValidateGenerationCreateRequestRejectsUnsupportedGPTImage2Size(t *testing.T) {
	req := CreateRequest{
		Kind:      "image",
		Provider:  "gpt",
		ModelCode: "gpt-image-2",
		Params:    map[string]any{"size": "1536x1024"},
	}
	if err := validateGenerationCreateRequest(req); err == nil {
		t.Fatal("expected unsupported 2K size to be rejected")
	}
}
