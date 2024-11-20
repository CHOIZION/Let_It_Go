package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Helper function to run a command
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Get the first line comment from a given file
func getFirstLineComment(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "//") {
			return strings.TrimSpace(line[2:]), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no comment found in the first line of %s", fileName)
}

func main() {
	// Initialize git repository
	fmt.Println("Initializing git repository...")
	if err := runCommand("git", "init"); err != nil {
		fmt.Printf("Failed to initialize git repository: %v\n", err)
		return
	}

	// Get current directory files
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Printf("Failed to read current directory: %v\n", err)
		return
	}

	// Find the first file with a comment
	var commitMessage string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") { // Assuming Go files, modify as needed
			commitMessage, err = getFirstLineComment(file.Name())
			if err == nil {
				fmt.Printf("Using commit message: '%s'\n", commitMessage)
				break
			}
		}
	}

	if commitMessage == "" {
		fmt.Println("No valid comment found for commit message. Exiting.")
		return
	}

	// Add all changes
	fmt.Println("Adding changes...")
	if err := runCommand("git", "add", "."); err != nil {
		fmt.Printf("Failed to add files: %v\n", err)
		return
	}

	// Commit changes with the message
	fmt.Println("Committing changes...")
	if err := runCommand("git", "commit", "-m", commitMessage); err != nil {
		fmt.Printf("Failed to commit changes: %v\n", err)
		return
	}

	// Push changes to origin/main
	fmt.Println("Pushing changes...")
	if err := runCommand("git", "push", "-u", "origin", "main"); err != nil {
		fmt.Printf("Failed to push changes: %v\n", err)
		return
	}

	fmt.Println("Git operations completed successfully!")
}
