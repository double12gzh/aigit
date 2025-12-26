package commit

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/zzxwill/aigit/internal/git"
	"github.com/zzxwill/aigit/internal/ui"
	"github.com/zzxwill/aigit/llm"
)

func SetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "Generate git commit message including title and body",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
}

func run() {
	diffOutput, err := git.GetStagedDiff()
	if err != nil {
		fmt.Printf("Error getting git diff: %v\n", err)
		os.Exit(1)
	}

	config := llm.NewConfig()
	if err := config.Load(); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	provider := getProvider(config)
	commitMessage, err := generateMessage(config, diffOutput)
	if err != nil {
		fmt.Printf("Error generating commit message: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🤖 Generating commit message by", provider)
	handleCommitLoop(config, diffOutput, commitMessage)
}

func getProvider(config *llm.Config) string {
	if config.CurrentProvider == "" {
		return llm.ProviderDoubao
	}
	return config.CurrentProvider
}

func handleCommitLoop(config *llm.Config, diffOutput []byte, initialMessage string) {
	commitMessage := initialMessage

	for {
		ui.DisplayCommitMessage(commitMessage)

		choice, err := ui.PromptCommitAction()
		if err != nil {
			fmt.Printf("Error with prompt: %v\n", err)
			os.Exit(1)
		}

		switch choice {
		case 0:
			if err := git.ExecuteCommit(commitMessage); err != nil {
				fmt.Printf("Error committing changes: %v\n", err)
				os.Exit(1)
			}
			if err := git.ExecutePush(); err != nil {
				color.Red("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case 1:
			var err error
			commitMessage, err = generateMessage(config, diffOutput)
			if err != nil {
				fmt.Printf("Error generating commit message: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("\n🤖 Regenerating commit message...")
		default:
			color.Red("Invalid choice")
		}
	}
}

func generateMessage(config *llm.Config, diffOutput []byte) (string, error) {
	generator, err := config.GetMessageGenerator()
	if err != nil {
		return "", fmt.Errorf("error getting message generator: %w", err)
	}
	return generator.GenerateCommitMessage(string(diffOutput))
}
