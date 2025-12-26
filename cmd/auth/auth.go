package auth

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/zzxwill/aigit/llm"
)

func SetupCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:                   "auth",
		Short:                 "Manage LLM providers and API keys",
		Long:                  `Manage Language Model providers and their API keys. Use subcommands to list, add, or select providers.`,
		DisableFlagsInUseLine: true,
	}

	authCmd.AddCommand(setupListCmd())
	authCmd.AddCommand(setupAddCmd())
	authCmd.AddCommand(setupUseCmd())

	return authCmd
}

func setupListCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "list",
		Aliases:               []string{"ls"},
		Short:                 "List configured LLM providers",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			runList()
		},
	}
}

func setupAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add or update API key for a provider",
		Long:  "Add or update API key for a provider. Supported providers: openai, gemini, doubao, deepseek, qwen, modelscope, self-hosted. Use --endpoint for Doubao and Self-hosted providers. Use --model-name for Modelscope and Self-hosted providers.",
		Run: func(cmd *cobra.Command, args []string) {
			runAdd(cmd)
		},
	}

	cmd.Flags().StringP("provider", "p", "", "LLM provider (openai, gemini, doubao, deepseek, qwen, modelscope, self-hosted)")
	cmd.Flags().StringP("api-key", "k", "", "API key for the provider")
	cmd.Flags().StringP("model-name", "m", "", "Model name (required for modelscope and self-hosted)")
	cmd.Flags().StringP("endpoint", "e", "", "Endpoint ID or URL (required for doubao and self-hosted)")
	if err := cmd.MarkFlagRequired("provider"); err != nil {
		panic(fmt.Sprintf("failed to mark provider flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("api-key"); err != nil {
		panic(fmt.Sprintf("failed to mark api-key flag as required: %v", err))
	}

	return cmd
}

func setupUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "use [provider]",
		Short:                 "Set the current LLM provider",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			runUse(args[0])
		},
	}
}

func runList() {
	config := llm.NewConfig()
	if err := config.Load(); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configured providers:")
	for _, provider := range config.ListProviders() {
		if provider == config.CurrentProvider {
			fmt.Printf("* %s (current)\n", provider)
		} else {
			fmt.Printf("  %s\n", provider)
		}
	}
}

func runAdd(cmd *cobra.Command) {
	provider, _ := cmd.Flags().GetString("provider")
	apiKey, _ := cmd.Flags().GetString("api-key")
	modelName, _ := cmd.Flags().GetString("model-name")
	endpoint, _ := cmd.Flags().GetString("endpoint")

	provider = strings.ToLower(provider)
	apiKey = strings.TrimSpace(apiKey)

	if provider == "" {
		color.Red("Provider is required")
		color.Red("Use --provider to specify the provider")
		os.Exit(1)
	}

	if apiKey == "" {
		color.Red("API key is required")
		color.Red("Use --api-key to specify the API key")
		os.Exit(1)
	}

	config := llm.NewConfig()
	if err := config.Load(); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	if err := addProvider(config, provider, apiKey, modelName, endpoint); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	color.Green("Successfully added API key for %s", provider)
}

func addProvider(config *llm.Config, provider, apiKey, modelName, endpoint string) error {
	switch provider {
	case llm.ProviderOpenAI, llm.ProviderGemini, llm.ProviderDeepseek, llm.ProviderQwen:
		return config.AddProvider(provider, apiKey, "")
	case llm.ProviderDoubao:
		if endpoint == "" {
			color.Red("Endpoint ID is required for Doubao provider")
			color.Red("Please run `aigit auth add --provider doubao --api-key <api_key> --endpoint <endpoint_id>`")
			return fmt.Errorf("endpoint ID is required for Doubao provider")
		}
		return config.AddProvider(provider, apiKey, "", endpoint)
	case llm.ProviderModelscope:
		if modelName == "" {
			color.Red("Model name is required for Modelscope provider")
			color.Red("Please run `aigit auth add --provider modelscope --api-key <api_key> --model-name <model_name>`")
			return fmt.Errorf("model name is required for Modelscope provider")
		}
		return config.AddProvider(provider, apiKey, modelName)
	case llm.ProviderSelfHosted:
		if modelName == "" {
			color.Red("Model name is required for Self-hosted provider")
			color.Red("Please run `aigit auth add --provider self-hosted --api-key <api_key> --model-name <model_name> --endpoint <endpoint>`")
			return fmt.Errorf("model name is required for Self-hosted provider")
		}
		if endpoint == "" {
			color.Red("Endpoint is required for Self-hosted provider")
			color.Red("Please run `aigit auth add --provider self-hosted --api-key <api_key> --model-name <model_name> --endpoint <endpoint>`")
			return fmt.Errorf("endpoint is required for Self-hosted provider")
		}
		return config.AddProvider(provider, apiKey, modelName, endpoint)
	default:
		return fmt.Errorf("unsupported provider: %s\nSupported providers are: openai, gemini, doubao, deepseek, qwen, modelscope, self-hosted", provider)
	}
}

func runUse(providerArg string) {
	provider := strings.ToLower(providerArg)

	config := llm.NewConfig()
	if err := config.Load(); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	if err := config.UseProvider(provider); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	color.Green("Now using %s as the current provider", provider)
}
