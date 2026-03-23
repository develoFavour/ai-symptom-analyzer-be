package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const groqAPIURL = "https://api.groq.com/openai/v1/chat/completions"
const groqModel = "llama-3.3-70b-versatile"

type groqClient struct {
	apiKey string
}

func NewGroqClient(apiKey string) Client {
	return &groqClient{apiKey: apiKey}
}

func (g *groqClient) Name() string { return "groq" }

func (g *groqClient) AnalyzeChat(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
	prompt := BuildSymptomPrompt(req)

	payload := map[string]interface{}{
		"model": groqModel,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a medical AI assistant. Always respond with valid JSON only. No markdown, no code blocks, just raw JSON.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("groq: marshal error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", groqAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("groq: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	rawBytes, err := doHTTPRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("groq: HTTP error: %w", err)
	}

	// Extract text from Groq/OpenAI-compatible response
	var groqResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(rawBytes, &groqResp); err != nil {
		return nil, fmt.Errorf("groq: response parse error: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("groq: empty response")
	}

	rawText := groqResp.Choices[0].Message.Content
	return ParseAIResponse(rawText)
}

func (g *groqClient) DefineMedicalTerm(ctx context.Context, term string) (string, error) {
	prompt := fmt.Sprintf("Provide a very concise, professional medical definition for the term: '%s'. Keep it under 2 sentences. No markdown.", term)

	payload := map[string]interface{}{
		"model": groqModel,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a medical dictionary assistant. Respond with the definition only.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.1,
		"max_tokens":  100,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("groq definition: marshal error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", groqAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("groq definition: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	rawBytes, err := doHTTPRequest(httpReq)
	if err != nil {
		return "", fmt.Errorf("groq definition: HTTP error: %w", err)
	}

	var groqResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rawBytes, &groqResp); err != nil {
		return "", fmt.Errorf("groq definition: response parse error: %w", err)
	}
	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("groq definition: empty response")
	}
	return groqResp.Choices[0].Message.Content, nil
}

func (g *groqClient) GenerateClinicalSummary(ctx context.Context, req SummarizeRequest) (string, error) {
	prompt := BuildSummaryPrompt(req)

	payload := map[string]interface{}{
		"model": groqModel,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a medical documentation assistant. Respond in plain text only — no JSON, no markdown.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.2,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("groq summary: marshal error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", groqAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("groq summary: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	rawBytes, err := doHTTPRequest(httpReq)
	if err != nil {
		return "", fmt.Errorf("groq summary: HTTP error: %w", err)
	}

	var groqResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rawBytes, &groqResp); err != nil {
		return "", fmt.Errorf("groq summary: response parse error: %w", err)
	}
	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("groq summary: empty response")
	}
	return groqResp.Choices[0].Message.Content, nil
}
