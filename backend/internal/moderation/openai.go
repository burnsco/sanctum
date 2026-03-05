// Package moderation provides content moderation utilities.
package moderation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Moderator checks text content for policy violations.
type Moderator interface {
	Check(ctx context.Context, text string) error
	CheckWithImage(ctx context.Context, text, imageURL string) error
}

// NoOp returns a Moderator that allows all content through (used when no API key is set).
func NoOp() Moderator { return noOpModerator{} }

type noOpModerator struct{}

func (noOpModerator) Check(_ context.Context, _ string) error              { return nil }
func (noOpModerator) CheckWithImage(_ context.Context, _, _ string) error  { return nil }

// OpenAI uses the OpenAI Moderation API (free) to screen text content.
// On API/network errors it fails open so a temporary outage doesn't block posts.
type OpenAI struct {
	apiKey string
	client *http.Client
}

// NewOpenAI returns an OpenAI moderator for the given API key.
func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// text-only request (default model)
type textModerationRequest struct {
	Input string `json:"input"`
}

// multimodal request (omni-moderation-latest supports image URLs)
type multimodalModerationRequest struct {
	Model string      `json:"model"`
	Input []inputItem `json:"input"`
}

type inputItem struct {
	Type     string     `json:"type"`
	Text     string     `json:"text,omitempty"`
	ImageURL *imageURLs `json:"image_url,omitempty"`
}

type imageURLs struct {
	URL string `json:"url"`
}

type moderationResponse struct {
	Results []struct {
		Flagged    bool            `json:"flagged"`
		Categories map[string]bool `json:"categories"`
	} `json:"results"`
}

const maxInputChars = 4096

// Check calls the OpenAI Moderation API and returns an error if the text is flagged.
func (m *OpenAI) Check(ctx context.Context, text string) error {
	return m.CheckWithImage(ctx, text, "")
}

// CheckWithImage calls the OpenAI Moderation API for text and optionally an image URL.
// When imageURL is non-empty, uses the omni-moderation-latest multimodal model.
func (m *OpenAI) CheckWithImage(ctx context.Context, text, imageURL string) error {
	if text == "" && imageURL == "" {
		return nil
	}

	var body []byte
	var err error

	if imageURL != "" {
		var items []inputItem
		if text != "" {
			if len(text) > maxInputChars {
				text = text[:maxInputChars]
			}
			items = append(items, inputItem{Type: "text", Text: text})
		}
		items = append(items, inputItem{Type: "image_url", ImageURL: &imageURLs{URL: imageURL}})
		body, err = json.Marshal(multimodalModerationRequest{
			Model: "omni-moderation-latest",
			Input: items,
		})
	} else {
		if len(text) > maxInputChars {
			text = text[:maxInputChars]
		}
		body, err = json.Marshal(textModerationRequest{Input: text})
	}

	if err != nil {
		return nil // fail open
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/moderations", bytes.NewReader(body))
	if err != nil {
		return nil // fail open
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		slog.WarnContext(ctx, "moderation API unreachable, failing open", "err", err)
		return nil // fail open on network error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.WarnContext(ctx, "moderation API returned non-200, failing open", "status", resp.StatusCode)
		return nil // fail open on API error
	}

	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil // fail open
	}

	var result moderationResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil // fail open
	}

	if len(result.Results) == 0 || !result.Results[0].Flagged {
		return nil
	}

	var triggered []string
	for cat, on := range result.Results[0].Categories {
		if on {
			triggered = append(triggered, strings.ReplaceAll(cat, "/", " "))
		}
	}
	if len(triggered) > 0 {
		return fmt.Errorf("content violates community guidelines (%s)", strings.Join(triggered, ", "))
	}
	return fmt.Errorf("content violates community guidelines")
}
