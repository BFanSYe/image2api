package gpt

import (
	"strings"
	"testing"

	"github.com/BFanSYe/image2api/backend/internal/provider"
)

func TestExtractWebImageToolIDs(t *testing.T) {
	raw := `{
		"conversation_id":"conv_123",
		"messages":[
			{
				"message":{
					"author":{"role":"tool"},
					"metadata":{"async_task_type":"image_gen"},
					"content":{
						"content_type":"multimodal_text",
						"parts":[
							{"asset_pointer":"file-service://file_abc12345"},
							{"asset_pointer":"sediment://sed_xyz98765"},
							"file-service://file_def67890"
						]
					}
				}
			}
		]
	}`
	fileIDs, sedimentIDs := extractWebImageToolIDs([]byte(raw))
	if len(fileIDs) != 2 {
		t.Fatalf("expected 2 file ids, got %v", fileIDs)
	}
	if len(sedimentIDs) != 1 || sedimentIDs[0] != "sed_xyz98765" {
		t.Fatalf("expected sediment id, got %v", sedimentIDs)
	}
}

func TestParseWebImageSSE(t *testing.T) {
	raw := strings.NewReader(strings.Join([]string{
		`data: {"type":"server_ste_metadata","conversation_id":"conv_456","metadata":{"tool_invoked":true,"turn_use_case":"image"}}`,
		"",
		`data: {"type":"response.completed","response":{"output":[{"type":"image_generation_call","result":"ZmFrZS1iNjQ=","output_format":"png"}]}}`,
		"",
	}, "\n"))

	conversationID, fileIDs, sedimentIDs, directURLs, lastText, err := parseWebImageSSE(raw)
	if err != nil {
		t.Fatalf("parseWebImageSSE error: %v", err)
	}
	if conversationID != "conv_456" {
		t.Fatalf("unexpected conversation id: %s", conversationID)
	}
	if len(fileIDs) != 0 || len(sedimentIDs) != 0 {
		t.Fatalf("unexpected ids: file=%v sediment=%v", fileIDs, sedimentIDs)
	}
	if len(directURLs) != 1 {
		t.Fatalf("expected 1 direct url, got %v", directURLs)
	}
	if lastText != "" {
		t.Fatalf("unexpected text: %q", lastText)
	}
}

func TestParseCompletedResponseFallsBackToLastPartialImage(t *testing.T) {
	raw := strings.NewReader(strings.Join([]string{
		`data: {"type":"response.image_generation_call.partial_image","partial_image_b64":"Zmlyc3Q=","output_format":"png"}`,
		"",
		`data: {"type":"response.image_generation_call.partial_image","partial_image_b64":"c2Vjb25k","output_format":"png"}`,
		"",
		`data: {"type":"response.completed","response":{"output":[{"type":"image_generation_call","output_format":"png"}]}}`,
		"",
	}, "\n"))

	completed, err := parseCompletedResponse(raw)
	if err != nil {
		t.Fatalf("parseCompletedResponse error: %v", err)
	}
	if len(completed.Response.Output) != 1 {
		t.Fatalf("expected 1 output, got %d", len(completed.Response.Output))
	}
	imageData, imageURL := outputImagePayload(completed.Response.Output[0])
	if imageURL != "" {
		t.Fatalf("unexpected image url: %s", imageURL)
	}
	if imageData != "c2Vjb25k" {
		t.Fatalf("expected last partial image, got %q", imageData)
	}
}

func TestParseWebImageSSEIgnoresUploadedReferenceIDs(t *testing.T) {
	raw := strings.NewReader(strings.Join([]string{
		`data: {"conversation_id":"conv_ref","message":{"author":{"role":"user"},"content":{"content_type":"multimodal_text","parts":[{"asset_pointer":"file-service://file_reference12345"},"make it transparent"]}}}`,
		"",
	}, "\n"))

	conversationID, fileIDs, sedimentIDs, directURLs, _, err := parseWebImageSSE(raw)
	if err != nil {
		t.Fatalf("parseWebImageSSE error: %v", err)
	}
	if conversationID != "conv_ref" {
		t.Fatalf("unexpected conversation id: %s", conversationID)
	}
	if len(fileIDs) != 0 || len(sedimentIDs) != 0 || len(directURLs) != 0 {
		t.Fatalf("reference image should not be treated as output: file=%v sediment=%v urls=%v", fileIDs, sedimentIDs, directURLs)
	}
}

func TestExtractWebImageToolIDsAcceptsAssistantImageMessages(t *testing.T) {
	raw := []byte(`{
		"conversation_id":"conv_assistant",
		"messages":[
			{
				"message":{
					"author":{"role":"assistant"},
					"metadata":{"async_task_type":"image_generation"},
					"content":{
						"content_type":"multimodal_text",
						"parts":[
							{"kind":"text","text":"done"},
							{"nested":{"asset_pointer":"file-service://file_out123456"}},
							["sediment://sed_out987654"]
						]
					}
				}
			}
		]
	}`)
	fileIDs, sedimentIDs := extractWebImageToolIDs(raw)
	if len(fileIDs) != 1 || fileIDs[0] != "file_out123456" {
		t.Fatalf("expected assistant image file id, got %v", fileIDs)
	}
	if len(sedimentIDs) != 1 || sedimentIDs[0] != "sed_out987654" {
		t.Fatalf("expected assistant sediment id, got %v", sedimentIDs)
	}
}

func TestWebImageMessageContentReferenceOrder(t *testing.T) {
	content, metadata := webImageMessageContent("make it transparent", []webUploadMeta{{
		FileID:        "file_ref123456",
		LibraryFileID: "libfile_ref123456",
		FileName:      "image_1.png",
		Mime:          "image/png",
		FileSize:      1234,
		Width:         512,
		Height:        512,
	}})

	if content["content_type"] != "multimodal_text" {
		t.Fatalf("unexpected content type: %v", content["content_type"])
	}
	parts, ok := content["parts"].([]any)
	if !ok || len(parts) != 2 {
		t.Fatalf("unexpected parts: %#v", content["parts"])
	}
	ref, ok := parts[0].(map[string]any)
	if !ok || ref["asset_pointer"] != "sediment://file_ref123456" {
		t.Fatalf("reference image should be first, got %#v", parts[0])
	}
	if parts[1] != "make it transparent" {
		t.Fatalf("prompt should be last, got %#v", parts[1])
	}
	attachments, ok := metadata["attachments"].([]map[string]any)
	if !ok || len(attachments) != 1 || attachments[0]["id"] != "file_ref123456" {
		t.Fatalf("unexpected attachments: %#v", metadata["attachments"])
	}
	if attachments[0]["source"] != "library" || attachments[0]["library_file_id"] != "libfile_ref123456" {
		t.Fatalf("unexpected attachment metadata: %#v", attachments[0])
	}
}

func TestExtractWebImageDirectURLsIgnoresChatGPTStaticAssets(t *testing.T) {
	raw := `{
		"url":"https://openaiassets.blob.core.windows.net/$web/chatgpt/filled-plus-icon.png",
		"image":"https://files.oaiusercontent.com/file-real-output.png?se=1"
	}`
	urls := extractWebImageDirectURLs(raw)
	if len(urls) != 1 || !strings.Contains(urls[0], "files.oaiusercontent.com") {
		t.Fatalf("expected only generated asset URL, got %#v", urls)
	}
}

func TestShouldUseWebImage2UsesNativeRouteForHighResolution(t *testing.T) {
	req := &provider.Request{Params: map[string]any{"resolution": "2K", "ratio": "16:9"}}
	if shouldUseWebImage2(req) {
		t.Fatal("expected native route for 2K")
	}
	if _, ok := req.Params["upscale_size"]; ok {
		t.Fatalf("native route must not add upscale_size: %#v", req.Params)
	}

	req = &provider.Request{Params: map[string]any{"size": "3312x1872"}}
	if shouldUseWebImage2(req) {
		t.Fatal("expected native route for explicit 4K size")
	}
	if _, ok := req.Params["upscale_method"]; ok {
		t.Fatalf("native route must not add upscale_method: %#v", req.Params)
	}
}

func TestShouldUseWebImage2KeepsWebRouteForOneK(t *testing.T) {
	req := &provider.Request{Params: map[string]any{"resolution": "1K", "ratio": "16:9"}}
	if !shouldUseWebImage2(req) {
		t.Fatal("expected web route for 1K")
	}
}

func TestNormalizeImage2NativeSizeKeepsUpstreamNativeSizes(t *testing.T) {
	if got := normalizeImage2NativeSize("3840x2160"); got != "3840x2160" {
		t.Fatalf("expected upstream native 4K size, got %s", got)
	}
	if got := normalizeImage2NativeSize("2048x2048"); got != "2048x2048" {
		t.Fatalf("expected upstream native 2K square size, got %s", got)
	}
	if got := normalizeImage2NativeSize("1664x928"); got != "2048x1152" {
		t.Fatalf("expected nearest upstream native 2K 16:9 size, got %s", got)
	}
}
