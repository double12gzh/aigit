package llm

import (
	"encoding/base64"
	"fmt"
)

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
