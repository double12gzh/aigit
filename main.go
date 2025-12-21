package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/zzxwill/aigit/llm"
)

var Version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "aigit",
		Short: "Generate git commit message including title and body",
		Long:  `AI Git Commi streamlines the git commit process by automatically generating meaningful and standardized commit messages.`,
	}

	authCmd := &cobra.Command{
		Use:                   "auth",
		Short:                 "Manage LLM providers and API keys",
		Long:                  `Manage Language Model providers and their API keys. Use subcommands to list, add, or select providers.`,
		DisableFlagsInUseLine: true,
	}

	authListCmd := &cobra.Command{
		Use:                   "list",
		Aliases:               []string{"ls"},
		Short:                 "List configured LLM providers",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
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
		},
	}

	authAddCmd := &cobra.Command{
		Use:                   "add <provider> <api_key> [endpoint_id]",
		Short:                 "Add or update API key for a provider",
		Long:                  "Add or update API key for a provider. Supported providers: openai, gemini, doubao, deepseek, qwen. endpoint_id is required for Doubao provider",
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				color.Red("Not enough arguments")
				color.Red(cmd.Long)
				color.Red("\nUsage: aigit auth add <provider> <api_key> [endpoint_id]")
				os.Exit(1)
			}

			provider := strings.ToLower(args[0])
			apiKey := strings.TrimSpace(args[1])

			config := llm.NewConfig()
			if err := config.Load(); err != nil {
				fmt.Printf("Error reading config: %v\n", err)
				os.Exit(1)
			}

			// Validate provider
			switch provider {
			case llm.ProviderOpenAI, llm.ProviderGemini, llm.ProviderDeepseek, llm.ProviderQwen:
				if err := config.AddProvider(provider, apiKey, ""); err != nil {
					fmt.Printf("Error saving config: %v\n", err)
					os.Exit(1)
				}
			case llm.ProviderDoubao:
				if len(args) < 3 {
					color.Red("Endpoint ID is required for Doubao provider")
					color.Red("Please run `aigit auth add doubao <api_key> <endpoint_id>`")
					os.Exit(1)
				}
				if err := config.AddProvider(provider, apiKey, "", args[2]); err != nil {
					fmt.Printf("Error saving config: %v\n", err)
					os.Exit(1)
				}
			case llm.ProviderModelscope:
				if len(args) < 2 {
					color.Red("Model name is required for Modelscope provider")
					color.Red("Please run `aigit auth add modelscope <api_key> <model_name>`")
					os.Exit(1)
				}
				if err := config.AddProvider(provider, apiKey, args[2]); err != nil {
					fmt.Printf("Error saving config: %v\n", err)
					os.Exit(1)
				}
			default:
				fmt.Printf("Unsupported provider: %s\nSupported providers are: openai, gemini, doubao, deepseek, qwen\n", provider)
				os.Exit(1)
			}

			color.Green("Successfully added API key for %s", provider)
		},
	}

	authUseCmd := &cobra.Command{
		Use:                   "use [provider]",
		Short:                 "Set the current LLM provider",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			provider := strings.ToLower(args[0])

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
		},
	}

	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authAddCmd)
	authCmd.AddCommand(authUseCmd)
	rootCmd.AddCommand(authCmd)

	commitCmd := &cobra.Command{
		Use:   "commit",
		Short: "Generate git commit message including title and body",
		Run: func(cmd *cobra.Command, args []string) {
			// Execute git diff --cached command
			diffOutput, err := exec.Command("git", "diff", "--cached").Output()
			if err != nil {
				fmt.Printf("Error getting git diff: %v\n", err)
				os.Exit(1)
			}

			// If there are no staged changes
			if len(diffOutput) == 0 {
				color.Yellow("No staged changes found.")
				stagePrompt := promptui.Select{
					Label: "Would you like to run 'git add .' to stage all changes?",
					Items: []string{"Yes", "No"},
					Size:  2,
				}

				_, stageChoice, err := stagePrompt.Run()
				if err != nil {
					fmt.Printf("Error with prompt: %v\n", err)
					os.Exit(1)
				}

				if stageChoice == "Yes" {
					cmd := exec.Command("git", "add", ".")
					if err := cmd.Run(); err != nil {
						fmt.Printf("Error staging changes: %v\n", err)
						os.Exit(1)
					}
					color.Green("All changes staged successfully!")

					// Re-run git diff to get the newly staged changes
					diffOutput, err = exec.Command("git", "diff", "--cached").Output()
					if err != nil {
						fmt.Printf("Error getting git diff: %v\n", err)
						os.Exit(1)
					}
				} else {
					color.Red("No changes staged. Please use 'git add' to stage your changes.")
					os.Exit(1)
				}
			}

			config := llm.NewConfig()
			if err := config.Load(); err != nil {
				fmt.Printf("Error reading config: %v\n", err)
				os.Exit(1)
			}

			var provider string
			if config.CurrentProvider == "" {
				provider = llm.ProviderDoubao
			} else {
				provider = config.CurrentProvider
			}

			// First message generation
			fmt.Println("\n🤖 Generating commit message by", provider)
			var commitMessage string
			commitMessage, err = generateMessage(config, diffOutput)
			if err != nil {
				fmt.Printf("Error generating commit message: %v\n", err)
				os.Exit(1)
			}

			for {
				// Clear some space and show the message in a box
				fmt.Println("\n📝 Generated commit message:")
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
				fmt.Println(commitMessage)
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

				fmt.Println("\n🤔 What would you like to do?")
				prompt := promptui.Select{
					Label: "Choose an action",
					Items: []string{"Commit this message", "Regenerate message"},
					Size:  2,
				}

				commitChoice, _, err := prompt.Run()
				if err != nil {
					fmt.Printf("Error with prompt: %v\n", err)
					os.Exit(1)
				}

				switch commitChoice {
				case 0:
					cmd := exec.Command("git", "commit", "-m", commitMessage)
					if err := cmd.Run(); err != nil {
						fmt.Printf("Error committing changes: %v\n", err)
						os.Exit(1)
					}
					color.Green("\n✅ Successfully committed changes!")

					pushPrompt := promptui.Select{
						Label: "Would you like to push these changes to the remote repository?",
						Items: []string{"No", "Yes"},
						Size:  2,
					}

					_, pushChoice, err := pushPrompt.Run()
					if err != nil {
						fmt.Printf("Error with prompt: %v\n", err)
						os.Exit(1)
					}

					if pushChoice == "Yes" {
						cmd := exec.Command("git", "push", "origin", "HEAD")
						output, err := cmd.CombinedOutput()
						if err != nil {
							color.Red("Error pushing changes: %v\n%s", err, output)
							os.Exit(1)
						}
						fmt.Printf("%s", output)
						color.Green("✅ Successfully pushed changes to remote repository!")
					} else {
						color.Yellow("Changes committed locally. Remember to push when ready.")
					}
					return
				case 1:
					fmt.Println("\n🤖 Regenerating commit message...")
					commitMessage, err = generateMessage(config, diffOutput)
					if err != nil {
						fmt.Printf("Error generating commit message: %v\n", err)
						os.Exit(1)
					}
					continue
				default:
					color.Red("Invalid choice")
				}
			}
		},
	}

	rootCmd.AddCommand(commitCmd)

	versionCmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v", "-v", "-version", "--version"},
		Short:   "Print the version of aigit",
		Long:    "Print the current version of the aigit CLI tool.",
		Run: func(cmd *cobra.Command, args []string) {
			if Version != "dev" {
				fmt.Println(Version)
				return
			}

			if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
				fmt.Println(info.Main.Version)
				return
			}

			version, err := exec.Command("git", "describe", "--tags").Output()
			if err != nil {
				fmt.Println("dev")
				return
			}
			fmt.Printf("%s\n", strings.TrimSpace(string(version)))
		},
	}

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateMessage(config *llm.Config, diffOutput []byte) (string, error) {
	generator, err := config.GetMessageGenerator()
	if err != nil {
		return "", fmt.Errorf("error getting message generator: %w", err)
	}
	return generator.GenerateCommitMessage(string(diffOutput))
}
