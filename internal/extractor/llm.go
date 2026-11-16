package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LLMConfig struct {
	Provider string
	Model    string
	APIKey   string
	Endpoint string
}

type OpenAIChatRequest struct {
	Model          string          `json:"model"`
	Messages       []ChatMessage   `json:"messages"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
	Temperature    float64         `json:"temperature"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type OpenAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ExtractWithLLM queries an OpenAI-compatible endpoint with strict JSON extraction.
func ExtractWithLLM(ctx context.Context, cfg LLMConfig, sender, subject, body string) (*ExtractedLead, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("no LLM API key configured")
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	systemPrompt := `You are an enterprise B2B lead extraction engine. Extract structured B2B lead attributes from the sales inquiry into pure JSON adhering strictly to this schema:
{
  "company_name": "Company Name",
  "domain": "canonical-domain.com",
  "contact_name": "Full Name",
  "job_title": "Job Title",
  "seniority": "C-Level / VP / Founder | Director / Head | Manager / Lead | Individual Contributor | Unspecified",
  "contact_email": "work-email@domain.com",
  "contact_phone": "+1...",
  "requested_products": ["Product 1"],
  "requested_seats": 100,
  "budget": 50000,
  "currency": "USD",
  "urgency": "P0 Immediate (< 1 mo) | P1 High (1-3 mos) | P2 Medium (3-6 mos) | P3 Exploratory (> 6 mos)",
  "tech_stack": ["Salesforce", "Okta"],
  "pain_points": "Summary of requirements"
}`

	userPrompt := fmt.Sprintf("Sender: %s\nSubject: %s\n\nContent:\n%s", sender, subject, body)

	reqPayload := OpenAIChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: &ResponseFormat{Type: "json_object"},
		Temperature:    0.1,
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(respData))
	}

	var chatResp OpenAIChatResponse
	if err := json.Unmarshal(respData, &chatResp); err != nil {
		return nil, fmt.Errorf("unmarshal chat response: %w", err)
	}

	if chatResp.Error != nil && chatResp.Error.Message != "" {
		return nil, fmt.Errorf("chat completion error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 || chatResp.Choices[0].Message.Content == "" {
		return nil, fmt.Errorf("empty response choices from LLM")
	}

	var lead ExtractedLead
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &lead); err != nil {
		return nil, fmt.Errorf("unmarshal extracted lead json: %w", err)
	}

	lead.ExtractionMethod = "llm"
	lead.ConfidenceScore = 0.95
	lead.Domain = NormalizeDomain(lead.Domain)

	return &lead, nil
}
