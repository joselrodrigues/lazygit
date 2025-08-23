package commit

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var CommitWithLlmErrorPrefix = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Handle LLM response with error prefix",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(appConfig *config.AppConfig) {
		// Configure LLM with a command that returns an error prefix
		appConfig.UserConfig.LLM.Enabled = true
		appConfig.UserConfig.LLM.Command = "echo 'error: API rate limit exceeded'"
	},
	SetupRepo: func(shell *Shell) {
		shell.
			CreateFile("test.txt", "initial content")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Lines(
				Contains("test.txt"),
			).
			PressPrimaryAction(). // stage the file
			Press(keys.Files.CommitChangesWithLLM)

		// Should show an error alert with the error message
		t.ExpectPopup().Alert().
			Title(Equals("Error")).
			Content(Contains("error: API rate limit exceeded")).
			Confirm()

		// Verify we're back in the files view
		t.Views().Files().
			IsFocused()

		// Verify no commits were created
		t.Views().Commits().
			IsEmpty()
	},
})