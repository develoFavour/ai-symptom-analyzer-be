package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

type geminiClient struct {
	apiKey string
}

func NewGeminiClient(apiKey string) Client {
	return &geminiClient{apiKey: apiKey}
}

func (g *geminiClient) Name() string { return "gemini" }

func (g *geminiClient) AnalyzeChat(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
	prompt := BuildSymptomPrompt(req)

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":      0.3, // Lower for more consistent medical responses
			"responseMimeType": "application/json",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshal error: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", geminiAPIURL, g.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("gemini: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	rawBytes, err := doHTTPRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: HTTP error: %w", err)
	}

	// Extract text from Gemini response structure
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(rawBytes, &geminiResp); err != nil {
		return nil, fmt.Errorf("gemini: response parse error: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini: empty response")
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text
	return ParseAIResponse(rawText)
}

func (g *geminiClient) DefineMedicalTerm(ctx context.Context, term string) (string, error) {
	prompt := fmt.Sprintf("Provide a very concise, professional medical definition for the term: '%s'. Keep it under 2 sentences. No markdown.", term)

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.1,
			"maxOutputTokens": 100,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("gemini definition: marshal error: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", geminiAPIURL, g.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("gemini definition: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	rawBytes, err := doHTTPRequest(httpReq)
	if err != nil {
		return "", fmt.Errorf("gemini definition: HTTP error: %w", err)
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(rawBytes, &geminiResp); err != nil {
		return "", fmt.Errorf("gemini definition: response parse error: %w", err)
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini definition: empty response")
	}
	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func (g *geminiClient) GenerateClinicalSummary(ctx context.Context, req SummarizeRequest) (string, error) {
	prompt := BuildSummaryPrompt(req)

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.2,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("gemini summary: marshal error: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", geminiAPIURL, g.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("gemini summary: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	rawBytes, err := doHTTPRequest(httpReq)
	if err != nil {
		return "", fmt.Errorf("gemini summary: HTTP error: %w", err)
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(rawBytes, &geminiResp); err != nil {
		return "", fmt.Errorf("gemini summary: response parse error: %w", err)
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini summary: empty response")
	}
	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
