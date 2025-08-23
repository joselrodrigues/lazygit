package commit

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var CommitWithLlmDisabled = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "LLM commit keybinding disabled when not configured",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(appConfig *config.AppConfig) {
		// LLM is disabled by default, but we'll explicitly set it
		appConfig.UserConfig.LLM.Enabled = false
		appConfig.UserConfig.LLM.Command = ""
	},
	SetupRepo: func(shell *Shell) {
		shell.
			CreateFile("test.txt", "content")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Lines(
				Contains("test.txt"),
			).
			PressPrimaryAction(). // stage the file
			Press(keys.Files.CommitChangesWithLLM)

		// Should show an error message or do nothing
		// The keybinding should be disabled, so pressing it should have no effect
		// We verify this by checking that no commit panel opens
		t.Views().Files().
			IsFocused() // Still in files view, no popup opened

		// Verify no commits were created
		t.Views().Commits().
			IsEmpty()
	},
})