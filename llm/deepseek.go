package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// DeepseekGenerator Implemention Deepseek provider
type DeepseekGenerator struct {
	apiKey string
}

func NewDeepseekGenerator(apiKey string) *DeepseekGenerator {
	return &DeepseekGenerator{apiKey: apiKey}
}

func (g *DeepseekGenerator) GenerateCommitMessage(diff string) (string, error) {
	return generateDeepseekCommitMessage(diff, g.apiKey)
}

func generateDeepseekCommitMessage(diff, apiKey string) (string, error) {
	client := &http.Client{}
	ctx := context.Background()
	prompt := fmt.Sprintf("%s\n%s", llmPrompt, diff)

	reqBody := map[string]interface{}{
		"model": deepseekModel,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are Deepseek, an AI assistant specialized in generating git commit messages.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream": false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content, nil
				}
			}
		}
	}

	return "", fmt.Errorf("invalid response format from Deepseek: %v", result)
}
