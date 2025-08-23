package helpers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"
)

// LLMHelper provides LLM-based commit message generation functionality
type LLMHelper struct {
	c *HelperCommon
}

// NewLLMHelper creates a new LLMHelper with injected dependencies
func NewLLMHelper(c *HelperCommon) *LLMHelper {
	return &LLMHelper{
		c: c,
	}
}


// GenerateCommitMessage generates a commit message using the configured LLM command.
// It validates the configuration, executes the external command with a timeout,
// and returns the generated commit message or an error if the process fails.
//
// The external command is expected to:
// - Analyze staged changes using git diff --cached
// - Output a properly formatted commit message to stdout
// - Return a non-zero exit code on error
func (self *LLMHelper) GenerateCommitMessage() (string, error) {
	// Validate LLM configuration
	llmConfig := self.c.UserConfig().LLM
	if !llmConfig.Enabled {
		return "", fmt.Errorf("%s", self.c.Tr.LLMDisabled)
	}
	if llmConfig.Command == "" {
		return "", fmt.Errorf("%s", self.c.Tr.LLMCommandNotConfigured)
	}

	// Log the execution for debugging
	self.c.Log.Debugf("Executing LLM command: %s", llmConfig.Command)

	// Create context with 30-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute the configured LLM command in a shell with timeout
	// The external script handles diff extraction and LLM communication
	cmd := exec.CommandContext(ctx, "sh", "-c", llmConfig.Command)
	
	output, err := cmd.Output()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("LLM command timed out after 30 seconds")
		}
		
		// Extract stderr if available for more informative error messages
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			errMsg := strings.TrimSpace(string(exitErr.Stderr))
			self.c.Log.Errorf("LLM command failed with stderr: %s", errMsg)
			return "", fmt.Errorf("LLM command failed: %s", errMsg)
		}
		
		self.c.Log.Errorf("LLM command execution failed: %v", err)
		return "", fmt.Errorf("failed to execute LLM command: %v", err)
	}

	// Clean the output
	message := strings.TrimSpace(string(output))
	
	// Validate the generated message
	if err := self.validateGeneratedMessage(message); err != nil {
		return "", err
	}

	// Handle error messages returned by the external script
	if strings.HasPrefix(message, "error:") {
		return "", fmt.Errorf("%s", message)
	}

	self.c.Log.Debugf("Successfully generated commit message: %d characters", len(message))
	return message, nil
}

// validateGeneratedMessage performs basic validation on the generated commit message
func (self *LLMHelper) validateGeneratedMessage(message string) error {
	// Check for empty message
	if message == "" {
		return fmt.Errorf("empty response from LLM command")
	}
	
	// Prevent obviously corrupted output (100KB sanity check)
	if len(message) > 100*1024 {
		return fmt.Errorf("generated message unreasonably large (%d bytes) - possible script error", len(message))
	}
	
	// Ensure valid UTF-8
	if !utf8.ValidString(message) {
		return fmt.Errorf("generated message contains invalid UTF-8 characters")
	}
	
	// Optional: Warn about format issues but don't fail
	lines := strings.Split(message, "\n")
	if len(lines) > 0 && len(lines[0]) > 72 {
		self.c.Log.Warnf("Generated commit subject exceeds 72 characters: %d", len(lines[0]))
	}
	
	return nil
}
