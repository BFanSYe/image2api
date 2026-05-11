package service

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/BFanSYe/image2api/backend/internal/model"
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
	for _, size := range []string{"1344x768", "1664x928", "2048x1152", "3312x1872", "3840x2160", "2160x3840"} {
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

func TestIsGPTImage2NativeAccountAllowsNativeUpstreamAccounts(t *testing.T) {
	apiKeyAccount := &model.Account{Provider: model.ProviderGPT, AuthType: model.AuthTypeAPIKey}
	if !isGPTImage2NativeAccount(apiKeyAccount) {
		t.Fatal("expected API key account to be eligible for native image route")
	}

	meta := `{"client_id":"` + codexOAuthClientID + `"}`
	codexOAuth := &model.Account{Provider: model.ProviderGPT, AuthType: model.AuthTypeOAuth, OAuthMeta: &meta}
	if !isGPTImage2NativeAccount(codexOAuth) {
		t.Fatal("expected Codex OAuth account to be eligible for native image route")
	}

	chatGPTBase := "https://chatgpt.com/backend-api"
	ordinaryOAuth := &model.Account{Provider: model.ProviderGPT, AuthType: model.AuthTypeOAuth, BaseURL: &chatGPTBase}
	if isGPTImage2NativeAccount(ordinaryOAuth) {
		t.Fatal("ordinary ChatGPT OAuth account without Codex client must not be used for native image route")
	}
}

func TestIsGPTImage2PreferredNativeAccountPrefersAPIKeyAndNativeBaseURL(t *testing.T) {
	apiKeyAccount := &model.Account{Provider: model.ProviderGPT, AuthType: model.AuthTypeAPIKey}
	if !isGPTImage2PreferredNativeAccount(apiKeyAccount) {
		t.Fatal("expected API key account to be preferred for native image route")
	}

	meta := `{"client_id":"` + codexOAuthClientID + `"}`
	codexOAuth := &model.Account{Provider: model.ProviderGPT, AuthType: model.AuthTypeOAuth, OAuthMeta: &meta}
	if isGPTImage2PreferredNativeAccount(codexOAuth) {
		t.Fatal("Codex OAuth account must be fallback-only when native API key accounts exist")
	}

	nativeBase := "https://api.example.test"
	nativeOAuth := &model.Account{Provider: model.ProviderGPT, AuthType: model.AuthTypeOAuth, BaseURL: &nativeBase}
	if !isGPTImage2PreferredNativeAccount(nativeOAuth) {
		t.Fatal("expected OAuth account with non-ChatGPT base URL to be preferred")
	}

	chatGPTBase := "https://chatgpt.com/backend-api"
	chatGPTAPIKey := &model.Account{Provider: model.ProviderGPT, AuthType: model.AuthTypeAPIKey, BaseURL: &chatGPTBase}
	if isGPTImage2PreferredNativeAccount(chatGPTAPIKey) {
		t.Fatal("ChatGPT base URL must not be preferred for native Responses image route")
	}
}

func TestShouldUseGPTWebRouteUsesNativeRouteForHighResolution(t *testing.T) {
	if shouldUseGPTWebRoute(map[string]any{"resolution": "2K", "ratio": "16:9"}) {
		t.Fatal("expected native route for 2K")
	}
	if shouldUseGPTWebRoute(map[string]any{"size": "3312x1872"}) {
		t.Fatal("expected native route for explicit 4K size")
	}
	if !shouldUseGPTWebRoute(map[string]any{"resolution": "1K", "ratio": "16:9"}) {
		t.Fatal("expected web route for 1K")
	}
}

func TestCacheDataURLAssetKeepsJPEGExtension(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("IMAGE2API_STORAGE_ROOT", tmp)

	dataURL := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString([]byte("not-a-real-jpeg"))
	s := &GenerationService{}
	url, ok := s.cacheDataURLAsset(context.Background(), "local", dataURL, "task02", 0, false)
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
