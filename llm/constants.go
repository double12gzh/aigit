package llm

const (
	ProviderOpenAI     = "openai"
	ProviderGemini     = "gemini"
	ProviderDoubao     = "doubao"
	ProviderDeepseek   = "deepseek"
	ProviderQwen       = "qwen"
	ProviderModelscope = "modelscope"
	ProviderSelfHosted = "self-hosted"

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

const llmPrompt = `You are a Git commit message generator. Generate a commit message that strictly follows Conventional Commits v1.0.0.

## Rules
- Format: <type>(<scope>): <subject>
- Scope is REQUIRED. Use a short noun (e.g., "auth", "ui", "config", "core"). If changes span many areas, use "core" or "app".
- Subject must be in imperative mood ("add", not "added" or "adds").
- Title and body (if any) must not exceed 72 characters per line.
- Separate title and body with a blank line.
- For breaking changes: append "!" after type (e.g., "feat!:"), OR include "BREAKING CHANGE: <description>" in the footer.
- Output ONLY the commit message—no explanations, prefixes, markdown, or reasoning tags (e.g., no <think>, no "Here is...", no code blocks).

## Types
feat: new feature
fix: bug fix
perf: performance improvement
refactor: code change with no behavior change
docs: documentation only
style: formatting, whitespace, etc.
test: test-related changes
build: build system or deps
ci: CI/CD changes
chore: maintenance tasks

## Output Format Example
feat(auth): add OAuth2 login support

Implement Google and GitHub OAuth providers to enable third-party login.

BREAKING CHANGE: removed legacy password-only auth flow

Now generate the commit message for the following diff:`
