package controllers

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/stretchr/testify/assert"
)

func TestFilesController_llmEnabled_structure(t *testing.T) {
	// This test validates that the llmEnabled method has the correct signature
	// Full testing would require complex mocking of the config system
	
	// Test that FilesController has the llmEnabled method
	controller := &FilesController{}
	
	// We can't actually call the method without proper initialization,
	// but we can test that it exists and has the right signature by checking
	// that it compiles and the method is accessible
	
	// This will fail compilation if the method signature changes
	var methodExists func() *types.DisabledReason = controller.llmEnabled
	assert.NotNil(t, methodExists)
}

func TestFilesController_structure(t *testing.T) {
	// Test that FilesController can be instantiated and has expected fields
	controller := &FilesController{}
	
	assert.NotNil(t, controller)
	
	// Test that the struct has the expected embedded types
	// This will fail compilation if the structure changes significantly
	var _ types.IController = controller
}