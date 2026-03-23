package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// ── Shared types ──────────────────────────────────────────────────────────────

// ChatMessage represents a single message in the conversational flow
type ChatMessage struct {
	Role    string `json:"role"` // "user" or "model"
	Content string `json:"content"`
}

// AnalysisRequest is the input sent to the AI for conversational triage
type AnalysisRequest struct {
	PatientContext string        `json:"patient_context"` // e.g., "Age: 30, Gender: Male"
	ChatHistory    []ChatMessage `json:"chat_history"`
	KnowledgeData  string        `json:"knowledge_data"` // stringified database records for RAG
}

// PossibleCondition represents a single diagnosis result
type PossibleCondition struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Confidence   string   `json:"confidence"` // high | medium | low
	CommonCauses []string `json:"common_causes"`
}

// AnalysisResult is the structured AI response
type AnalysisResult struct {
	IsDiagnosisReady   bool                `json:"is_diagnosis_ready"` // false if asking follow-up questions
	BotMessage         string              `json:"bot_message"`        // strictly follow-up questions if not ready, or a polite preamble if ready
	PossibleConditions []PossibleCondition `json:"possible_conditions,omitempty"`
	UrgencyLevel       string              `json:"urgency_level,omitempty"` // emergency | see_doctor | self_care
	HealthAdvice       string              `json:"health_advice,omitempty"`
	Disclaimer         string              `json:"disclaimer,omitempty"`
}

// SummarizeRequest is used to generate a clinical summary of a chat session
type SummarizeRequest struct {
	ChatHistory    []ChatMessage
	PatientContext string
}

// Client is the interface both Gemini and Groq implement
type Client interface {
	AnalyzeChat(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error)
	GenerateClinicalSummary(ctx context.Context, req SummarizeRequest) (string, error)
	DefineMedicalTerm(ctx context.Context, term string) (string, error)
	Name() string
}

// ── Resilient AI client (primary + fallback) ──────────────────────────────────

type resilientClient struct {
	primary  Client
	fallback Client
}

// NewResilientClient returns an AI client that falls back to Groq if Gemini fails
func NewResilientClient() Client {
	return &resilientClient{
		primary:  NewGeminiClient(os.Getenv("GEMINI_API_KEY")),
		fallback: NewGroqClient(os.Getenv("GROQ_API_KEY")),
	}
}

func (r *resilientClient) Name() string {
	return "resilient(" + r.primary.Name() + "/" + r.fallback.Name() + ")"
}

func (r *resilientClient) AnalyzeChat(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
	result, err := r.primary.AnalyzeChat(ctx, req)
	if err != nil {
		// Primary failed — log the error and fallback to Groq
		log.Printf("[resilientClient] %s failed: %v. Falling back to %s...", r.primary.Name(), err, r.fallback.Name())
		fallbackResult, fallbackErr := r.fallback.AnalyzeChat(ctx, req)
		if fallbackErr != nil {
			return nil, fmt.Errorf("primary failed: %v | fallback also failed: %v", err, fallbackErr)
		}
		return fallbackResult, nil
	}
	return result, nil
}

func (r *resilientClient) GenerateClinicalSummary(ctx context.Context, req SummarizeRequest) (string, error) {
	summary, err := r.primary.GenerateClinicalSummary(ctx, req)
	if err != nil {
		log.Printf("[resilientClient] %s summary failed: %v. Falling back to %s...", r.primary.Name(), err, r.fallback.Name())
		return r.fallback.GenerateClinicalSummary(ctx, req)
	}
	return summary, nil
}

func (r *resilientClient) DefineMedicalTerm(ctx context.Context, term string) (string, error) {
	definition, err := r.primary.DefineMedicalTerm(ctx, term)
	if err != nil {
		log.Printf("[resilientClient] %s definition failed: %v. Falling back to %s...", r.primary.Name(), err, r.fallback.Name())
		return r.fallback.DefineMedicalTerm(ctx, term)
	}
	return definition, nil
}

// ── Shared prompt builders ─────────────────────────────────────────────────────

// BuildSummaryPrompt creates a medical summarization prompt for a chat session
func BuildSummaryPrompt(req SummarizeRequest) string {
	history := ""
	for _, msg := range req.ChatHistory {
		role := "Patient"
		if msg.Role == "model" {
			role = "Vitalis AI"
		}
		history += fmt.Sprintf("%s: %s\n", role, msg.Content)
	}
	return fmt.Sprintf(`You are a medical documentation assistant. 
A patient had the following AI-assisted triage conversation. 
Generate a concise, structured CLINICAL SUMMARY that a doctor can read in under one minute.

Patient Context: %s

Conversation:
%s

Output format (plain text, no JSON, no markdown headers):
- Patient's Complaint: [main symptom in one line]
- Symptom Details: [duration, severity, triggers, associated symptoms — 2-3 sentences]
- AI Assessment: [conditions considered and urgency level]
- Patient's Question/Note: [what the patient most wants addressed]
- Recommended Focus: [what the doctor should pay attention to]

Keep it professional, clinical, and under 200 words.`, req.PatientContext, history)
}

func BuildSymptomPrompt(req AnalysisRequest) string {
	history := ""
	userTurnCount := 0
	for _, msg := range req.ChatHistory {
		role := "User"
		if msg.Role == "model" {
			role = "Vitalis (Assistant)"
		} else {
			userTurnCount++
		}
		history += fmt.Sprintf("%s: %s\n", role, msg.Content)
	}

	// Determine whether we are forcing a diagnosis
	forceDecision := ""
	if userTurnCount >= 3 {
		forceDecision = fmt.Sprintf(`
IMPORTANT OVERRIDE — The patient has already responded %d times. 
You have gathered enough information for a preliminary assessment. 
You MUST now set "is_diagnosis_ready" to true and provide your best possible assessment based on ALL the details shared so far.
DO NOT ask any more follow-up questions. Make a decision now using the Provided Medical Dataset.`, userTurnCount)
	}

	return fmt.Sprintf(`You are Vitalis AI, a concise and empathetic medical triage assistant.
You are chatting with a patient. Your job is to gather just enough details to make a safe preliminary assessment, then commit to it.

Patient Context:
%s

Provided Medical Dataset (from our approved database — your ONLY source for diagnoses):
%s

Conversation History (%d patient replies so far):
%s
%s
Rules:
1. CLARIFICATION PHASE (Patient has replied fewer than 3 times): You MUST set "is_diagnosis_ready" to false and ask 1-2 short, focused follow-up questions to rule out other conditions (e.g., "How long have you had this?", "Any other symptoms?"). DO NOT diagnose yet, even if you see a strong match in the Medical Dataset.
2. DIAGNOSIS PHASE (Patient has replied 3+ times): Set "is_diagnosis_ready" to true. Provide a warm "bot_message" summarizing your assessment, then populate "possible_conditions", "urgency_level", and "health_advice" strictly from the Provided Medical Dataset.
3. SOURCE CITATIONS: When giving a diagnosis/advice (Phase 2), you MUST humanely cite the source from the Medical Dataset in your "bot_message" (e.g., "According to WHO/NCDC protocols...", "Following Mayo Clinic guidelines...").
4. NEVER repeat a question. Read the conversation history carefully.
5. "urgency_level" must be exactly one of: emergency, see_doctor, self_care.
6. "confidence" must be exactly one of: high, medium, low.
7. Always include a "disclaimer".
8. Base ALL diagnoses and advice STRICTLY on the Provided Medical Dataset. Do not invent conditions.

Respond ONLY with a valid JSON object matching this exact schema (no markdown, no extra text):
{
  "is_diagnosis_ready": boolean,
  "bot_message": "Your friendly conversational reply to the patient",
  "possible_conditions": [
    {
      "name": "Condition Name",
      "description": "Brief description",
      "confidence": "high|medium|low",
      "common_causes": ["cause1", "cause2"]
    }
  ],
  "urgency_level": "emergency|see_doctor|self_care",
  "health_advice": "Clear, actionable advice based on symptoms",
  "disclaimer": "This is a preliminary assessment. Please consult a qualified healthcare professional."
}
`, req.PatientContext, req.KnowledgeData, userTurnCount, history, forceDecision)
}

// ParseAIResponse parses the JSON returned by AI into AnalysisResult
func ParseAIResponse(raw string) (*AnalysisResult, error) {
	// Clean markdown if present
	raw = strings.TrimPrefix(strings.TrimSpace(raw), "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")

	var result AnalysisResult
	if err := json.NewDecoder(bytes.NewBufferString(raw)).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w\nRaw string: %s", err, raw)
	}
	return &result, nil
}

// doHTTPRequest is a shared HTTP helper for AI API calls
func doHTTPRequest(req *http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("AI API returned status %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	return buf.Bytes(), err
}
