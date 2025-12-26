package git

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"

	"github.com/zzxwill/aigit/internal/ui"
)

func GetStagedDiff() ([]byte, error) {
	diffOutput, err := exec.Command("git", "diff", "--cached").Output()
	if err != nil {
		return nil, err
	}

	if len(diffOutput) == 0 {
		if err := promptAndStage(); err != nil {
			return nil, err
		}
		diffOutput, err = exec.Command("git", "diff", "--cached").Output()
		if err != nil {
			return nil, err
		}
	}

	return diffOutput, nil
}

func promptAndStage() error {
	color.Yellow("No staged changes found.")
	choice, err := ui.PromptStageChanges()
	if err != nil {
		return fmt.Errorf("error with prompt: %w", err)
	}

	if choice == "Yes" {
		cmd := exec.Command("git", "add", ".")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error staging changes: %w", err)
		}
		color.Green("All changes staged successfully!")
		return nil
	}

	color.Red("No changes staged. Please use 'git add' to stage your changes.")
	os.Exit(1)
	return nil
}

func ExecuteCommit(commitMessage string) error {
	cmd := exec.Command("git", "commit", "-m", commitMessage)
	if err := cmd.Run(); err != nil {
		return err
	}
	color.Green("\n✅ Successfully committed changes!")
	return nil
}

func ExecutePush() error {
	choice, err := ui.PromptPush()
	if err != nil {
		return fmt.Errorf("error with prompt: %w", err)
	}

	if choice == "Yes" {
		cmd := exec.Command("git", "push", "origin", "HEAD")
		output, err := cmd.CombinedOutput()
		if err != nil {
			color.Red("Error pushing changes: %v\n%s", err, output)
			return fmt.Errorf("error pushing changes: %w", err)
		}
		fmt.Printf("%s", output)
		color.Green("✅ Successfully pushed changes to remote repository!")
	} else {
		color.Yellow("Changes committed locally. Remember to push when ready.")
	}

	return nil
}
