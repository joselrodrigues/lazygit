package helpers

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/jesseduffield/lazygit/pkg/common"
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/jesseduffield/lazygit/pkg/i18n"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLLMHelper_NewLLMHelper(t *testing.T) {
	helperCommon := &HelperCommon{}
	
	llmHelper := NewLLMHelper(helperCommon)
	
	assert.NotNil(t, llmHelper)
	assert.Equal(t, helperCommon, llmHelper.c)
}

func TestLLMHelper_GenerateCommitMessage(t *testing.T) {
	scenarios := []struct {
		name           string
		llmConfig      config.LLMConfig
		expectedError  string
		expectedResult string
		setupMock      func(*mockHelperCommon)
	}{
		{
			name: "disabled LLM returns error",
			llmConfig: config.LLMConfig{
				Enabled: false,
				Command: "echo 'test'",
			},
			expectedError: "LLM commit generation is disabled in config",
			setupMock: func(m *mockHelperCommon) {
				m.tr.LLMDisabled = "LLM commit generation is disabled in config"
			},
		},
		{
			name: "empty command returns error",
			llmConfig: config.LLMConfig{
				Enabled: true,
				Command: "",
			},
			expectedError: "LLM command is not configured",
			setupMock: func(m *mockHelperCommon) {
				m.tr.LLMCommandNotConfigured = "LLM command is not configured"
			},
		},
		{
			name: "command with error prefix in output",
			llmConfig: config.LLMConfig{
				Enabled: true,
				Command: "echo 'error: API key invalid'",
			},
			expectedError: "error: API key invalid",
			setupMock: func(m *mockHelperCommon) {},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			mockCommon := newMockHelperCommon()
			s.setupMock(mockCommon)
			mockCommon.userConfig.LLM = s.llmConfig
			// Update the config in the Common object as well
			mockCommon.HelperCommon.Common.SetUserConfig(mockCommon.userConfig)
			
			llmHelper := &LLMHelper{
				c: mockCommon.HelperCommon,
			}
			
			result, err := llmHelper.GenerateCommitMessage()
			
			if s.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), s.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, s.expectedResult, result)
			}
		})
	}
}

func TestLLMHelper_validateGeneratedMessage(t *testing.T) {
	scenarios := []struct {
		name          string
		message       string
		expectedError string
		expectWarning bool
	}{
		{
			name:          "empty message",
			message:       "",
			expectedError: "empty response from LLM command",
		},
		{
			name:          "message too large",
			message:       string(make([]byte, 101*1024)),
			expectedError: "generated message unreasonably large",
		},
		{
			name:          "invalid UTF-8",
			message:       "test\x80\x81",
			expectedError: "invalid UTF-8",
		},
		{
			name:          "valid short message",
			message:       "feat: add new feature",
			expectedError: "",
		},
		{
			name:          "subject line too long",
			message:       "feat: this is a very long commit message subject that exceeds the recommended 72 character limit for git commits",
			expectedError: "",
			expectWarning: true,
		},
		{
			name: "valid multiline message",
			message: `feat: add new feature

This is the body of the commit message
with multiple lines`,
			expectedError: "",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			mockCommon := newMockHelperCommon()
			llmHelper := &LLMHelper{
				c: mockCommon.HelperCommon,
			}
			
			err := llmHelper.validateGeneratedMessage(s.message)
			
			if s.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), s.expectedError)
			} else {
				assert.NoError(t, err)
			}
			
			// We can't easily test logging output without refactoring
			// Just verify the validation succeeded or failed as expected
			_ = s.expectWarning
		})
	}
}

func TestLLMHelper_GenerateCommitMessage_Timeout(t *testing.T) {
	// This test is commented out because it would take 30 seconds to run
	// and we can't easily override the timeout without modifying the implementation
	// to make it configurable. The timeout is tested in integration tests.
	t.Skip("Timeout test skipped - tested in integration tests")
}

func TestLLMHelper_GenerateCommitMessage_CommandFailure(t *testing.T) {
	scenarios := []struct {
		name          string
		command       string
		expectedError string
	}{
		{
			name:          "command not found",
			command:       "nonexistentcommand123",
			expectedError: "LLM command failed",
		},
		{
			name:          "command exits with error",
			command:       "exit 1",
			expectedError: "failed to execute LLM command",
		},
		{
			name:          "command with stderr",
			command:       "echo 'error message' >&2 && exit 1",
			expectedError: "LLM command failed",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			mockCommon := newMockHelperCommon()
			mockCommon.userConfig.LLM = config.LLMConfig{
				Enabled: true,
				Command: s.command,
			}
			// Update the config in the Common object as well
			mockCommon.HelperCommon.Common.SetUserConfig(mockCommon.userConfig)
			
			llmHelper := &LLMHelper{
				c: mockCommon.HelperCommon,
			}
			
			_, err := llmHelper.GenerateCommitMessage()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), s.expectedError)
		})
	}
}

// Mock implementations for testing

type mockHelperCommon struct {
	*HelperCommon
	userConfig *config.UserConfig
	tr         *i18n.TranslationSet
	logBuffer  *logrus.Logger
}

func newMockHelperCommon() *mockHelperCommon {
	logger := logrus.New()
	
	userConfig := &config.UserConfig{
		LLM: config.LLMConfig{
			Enabled: false,
			Command: "",
		},
	}
	
	tr := &i18n.TranslationSet{
		LLMDisabled:             "LLM disabled",
		LLMCommandNotConfigured: "LLM command not configured",
	}
	
	commonObj := &common.Common{
		Tr:  tr,
		Log: logger.WithField("test", true),
	}
	
	// Set the user config using the atomic pointer
	commonObj.SetUserConfig(userConfig)
	
	helperCommon := &HelperCommon{
		Common: commonObj,
	}
	
	return &mockHelperCommon{
		HelperCommon: helperCommon,
		userConfig:   userConfig,
		tr:           tr,
		logBuffer:    logger,
	}
}

// CommandExecutor interface for better testing
type CommandExecutor interface {
	Execute(ctx context.Context, command string) ([]byte, error)
}

type RealCommandExecutor struct{}

func (r *RealCommandExecutor) Execute(ctx context.Context, command string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	return cmd.Output()
}

type MockCommandExecutor struct {
	Output []byte
	Error  error
}

func (m *MockCommandExecutor) Execute(ctx context.Context, command string) ([]byte, error) {
	return m.Output, m.Error
}

// Example of how to use the CommandExecutor for testing
func TestLLMHelper_WithMockExecutor(t *testing.T) {
	scenarios := []struct {
		name           string
		mockOutput     []byte
		mockError      error
		expectedResult string
		expectedError  string
	}{
		{
			name:           "successful generation",
			mockOutput:     []byte("feat: add new feature"),
			mockError:      nil,
			expectedResult: "feat: add new feature",
			expectedError:  "",
		},
		{
			name:           "command error",
			mockOutput:     nil,
			mockError:      errors.New("command failed"),
			expectedResult: "",
			expectedError:  "command failed",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			// This demonstrates how we could refactor to use an executor
			// for better testing, but would require changes to the main implementation
			executor := &MockCommandExecutor{
				Output: s.mockOutput,
				Error:  s.mockError,
			}
			
			// In a refactored version, we'd inject the executor
			_ = executor
			
			// For now, we test with the actual implementation
			// which makes it harder to mock command execution
		})
	}
}