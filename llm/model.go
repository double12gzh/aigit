package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/openai/openai-go"
	openaioption "github.com/openai/openai-go/option"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"google.golang.org/api/option"
)

const (
	ProviderOpenAI     = "openai"
	ProviderGemini     = "gemini"
	ProviderDoubao     = "doubao"
	ProviderDeepseek   = "deepseek"
	ProviderQwen       = "qwen"
	ProviderModelscope = "modelscope"

	// Model constants
	geminiModel   = "gemini-pro"
	deepseekModel = "deepseek-chat"
	openaiModel   = "chatgpt-4o-latest"
	qwenModel     = "qwen-plus"
)

const (
	DefaultAPIKey   = "YzAzZDIyN2ItMWFkNS00MDNkLWJkM2YtZjgzNzczOWE4YzFj"
	DefaultEndpoint = "ZXAtMjAyNTAxMTMyMzE5NTEtOTJ4bjI="
)

const llmPrompt = `Generate a Git commit message following Conventional Commits v1.0.0: <type>(<scope>): <description>

# Type Selection (by priority):
BREAKING CHANGE: Add ! after type or BREAKING CHANGE: in footer for API changes
fix: Bug fixes, crashes, errors, security issues
feat: New features, APIs, capabilities
perf: Performance improvements
refactor: Code restructuring without functional changes
docs: Documentation only
style: Formatting, whitespace, imports
test: Test changes
build: Build system, dependencies
ci: CI/CD changes
chore: Maintenance, version updates

The commit message should follow these rules:
1. Follow the Conventional Commits format: <type>(<scope>): <description>
2. The body should be one paragraph
3. The body should explain WHAT and WHY (not HOW)
4. Each line should be less than 72 characters
5. There should be a line break between the title and the body
6. Disable markdown formatting

IMPORTANT: Only output the commit message itself, no additional explanation, no template description, no prefatory text. Just the commit message in the format specified.

Here's the diff:`

var (
	_ MessageGenerator = (*DefaultGenerator)(nil)
	_ MessageGenerator = (*GeminiGenerator)(nil)
	_ MessageGenerator = (*OpenAIGenerator)(nil)
	_ MessageGenerator = (*DoubaoGenerator)(nil)
	_ MessageGenerator = (*DeepseekGenerator)(nil)
	_ MessageGenerator = (*QwenGenerator)(nil)
	_ MessageGenerator = (*ModelscopeGenerator)(nil)
)

// MessageGenerator Define a commit message generator
type MessageGenerator interface {
	GenerateCommitMessage(diff string) (string, error)
}

type DefaultGenerator struct {
	MessageGenerator MessageGenerator
}

func NewDefauleGenerator() (*DefaultGenerator, error) {
	apiKey, err := base64.StdEncoding.DecodeString(DefaultAPIKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding API key: %w", err)
	}
	endpoint, err := base64.StdEncoding.DecodeString(DefaultEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error decoding endpoint: %w", err)
	}
	generator := &DefaultGenerator{
		MessageGenerator: NewDoubaoGenerator(string(apiKey), string(endpoint)),
	}

	return generator, nil
}

func (d *DefaultGenerator) GenerateCommitMessage(diff string) (string, error) {
	return d.MessageGenerator.GenerateCommitMessage(diff)
}

// GeminiGenerator Implemention Gemini provider
type GeminiGenerator struct {
	apiKey string
}

func NewGeminiGenerator(apiKey string) *GeminiGenerator {
	return &GeminiGenerator{apiKey: apiKey}
}

func (g *GeminiGenerator) GenerateCommitMessage(diff string) (string, error) {
	return generateGeminiCommitMessage(diff, g.apiKey)
}

// OpenAIGenerator Implemention OpenAI provider
type OpenAIGenerator struct {
	apiKey string
}

func NewOpenAIGenerator(apiKey string) *OpenAIGenerator {
	return &OpenAIGenerator{apiKey: apiKey}
}

func (g *OpenAIGenerator) GenerateCommitMessage(diff string) (string, error) {
	return generateOpenAICommitMessage(diff, g.apiKey)
}

// DoubaoGenerator Implemention Doubao provider
type DoubaoGenerator struct {
	apiKey   string
	endpoint string
}

func NewDoubaoGenerator(apiKey, endpoint string) *DoubaoGenerator {
	return &DoubaoGenerator{apiKey: apiKey, endpoint: endpoint}
}

func (g *DoubaoGenerator) GenerateCommitMessage(diff string) (string, error) {
	return generateDoubaoCommitMessage(diff, g.apiKey, g.endpoint)
}

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

// QwenGenerator Implemention Qwen provider
type QwenGenerator struct {
	apiKey string
}

func NewQwenGenerator(apiKey string) *QwenGenerator {
	return &QwenGenerator{apiKey: apiKey}
}

func (g *QwenGenerator) GenerateCommitMessage(diff string) (string, error) {
	return generateQwenCommitMessage(diff, g.apiKey, qwenModel, "https://dashscope.aliyuncs.com/compatible-mode/v1/")
}

// ModelscopeGenerator Implemention Modelscope provider
type ModelscopeGenerator struct {
	apiKey    string
	modelName string
}

func NewModelscopeGenerator(apiKey, modelName string) *ModelscopeGenerator {
	return &ModelscopeGenerator{apiKey: apiKey, modelName: modelName}
}

func (g *ModelscopeGenerator) GenerateCommitMessage(diff string) (string, error) {
	return generateQwenCommitMessage(diff, g.apiKey, g.modelName, "https://api-inference.modelscope.cn/v1/")
}

func generateGeminiCommitMessage(diff, apiKey string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("creating Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(geminiModel)
	prompt := fmt.Sprintf("%s\n%s", llmPrompt, diff)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("generating commit message: %w", err)
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		generatedMessage := resp.Candidates[0].Content.Parts[0].(genai.Text)
		return strings.TrimSpace(string(generatedMessage)), nil
	}

	return "", fmt.Errorf("no commit message generated by Gemini")
}

func generateOpenAICommitMessage(diff, apiKey string) (string, error) {
	client := openai.NewClient(
		openaioption.WithAPIKey(apiKey),
	)
	ctx := context.Background()
	prompt := fmt.Sprintf("%s\n%s", llmPrompt, diff)

	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		return "", fmt.Errorf("generating commit message: %w", err)
	}

	if len(chatCompletion.Choices) > 0 {
		return strings.TrimSpace(chatCompletion.Choices[0].Message.Content), nil
	}

	return "", fmt.Errorf("no commit message generated by OpenAI")
}

func generateDoubaoCommitMessage(diff, apiKey string, endpointId string) (string, error) {
	client := arkruntime.NewClientWithApiKey(
		apiKey,
	)

	ctx := context.Background()

	prompt := fmt.Sprintf("%s\n%s", llmPrompt, diff)

	req := model.ChatCompletionRequest{
		Model: endpointId,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleSystem,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String("你是豆包，是由字节跳动开发的 AI 人工智能助手，你非常擅长生成 git commit message"),
				},
			},
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(prompt),
				},
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return *resp.Choices[0].Message.Content.StringValue, nil
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

func generateQwenCommitMessage(diff, apiKey string, modelName string, baseURL string) (string, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		apiKey = os.Getenv("DASHSCOPE_API_KEY")
	}

	if apiKey == "" {
		return "", fmt.Errorf("API key is empty: please provide an API key via configuration or DASHSCOPE_API_KEY environment variable")
	}

	client := openai.NewClient(
		openaioption.WithAPIKey(apiKey),
		openaioption.WithBaseURL(baseURL),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	prompt := fmt.Sprintf("%s\n%s", llmPrompt, diff)

	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		Model: openai.F(modelName),
	})
	if err != nil {
		return "", fmt.Errorf("generating commit message: %w", err)
	}

	if len(chatCompletion.Choices) > 0 {
		return strings.TrimSpace(chatCompletion.Choices[0].Message.Content), nil
	}

	return "", fmt.Errorf("no commit message generated by Qwen")
}
