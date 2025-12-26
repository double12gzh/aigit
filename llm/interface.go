package llm

// MessageGenerator Define a commit message generator
type MessageGenerator interface {
	GenerateCommitMessage(diff string) (string, error)
}

var (
	_ MessageGenerator = (*DefaultGenerator)(nil)
	_ MessageGenerator = (*GeminiGenerator)(nil)
	_ MessageGenerator = (*OpenAIGenerator)(nil)
	_ MessageGenerator = (*DoubaoGenerator)(nil)
	_ MessageGenerator = (*DeepseekGenerator)(nil)
	_ MessageGenerator = (*QwenGenerator)(nil)
	_ MessageGenerator = (*ModelscopeGenerator)(nil)
	_ MessageGenerator = (*SelfHostedGenerator)(nil)
)
