package llm

import (
	"context"
	"fmt"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

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

func generateDoubaoCommitMessage(diff, apiKey string, endpointID string) (string, error) {
	client := arkruntime.NewClientWithApiKey(
		apiKey,
	)

	ctx := context.Background()

	prompt := fmt.Sprintf("%s\n%s", llmPrompt, diff)

	req := model.ChatCompletionRequest{
		Model: endpointID,
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
