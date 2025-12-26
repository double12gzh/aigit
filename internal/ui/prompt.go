package ui

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

func PromptStageChanges() (string, error) {
	prompt := promptui.Select{
		Label: "Would you like to run 'git add .' to stage all changes?",
		Items: []string{"Yes", "No"},
		Size:  2,
	}

	_, choice, err := prompt.Run()
	return choice, err
}

func PromptCommitAction() (int, error) {
	fmt.Println("\n🤔 What would you like to do?")
	prompt := promptui.Select{
		Label: "Choose an action",
		Items: []string{"Commit this message", "Regenerate message"},
		Size:  2,
	}

	choice, _, err := prompt.Run()
	return choice, err
}

func PromptPush() (string, error) {
	prompt := promptui.Select{
		Label: "Would you like to push these changes to the remote repository?",
		Items: []string{"No", "Yes"},
		Size:  2,
	}

	_, choice, err := prompt.Run()
	return choice, err
}
