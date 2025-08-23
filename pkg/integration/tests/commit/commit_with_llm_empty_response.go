package commit

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var CommitWithLlmEmptyResponse = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Handle empty LLM response gracefully",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(appConfig *config.AppConfig) {
		// Configure LLM with a command that returns empty output
		appConfig.UserConfig.LLM.Enabled = true
		appConfig.UserConfig.LLM.Command = "echo ''"
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

		// Should show an error alert about empty response
		t.ExpectPopup().Alert().
			Title(Equals("Error")).
			Content(Contains("empty response")).
			Confirm()

		// Verify we're back in the files view
		t.Views().Files().
			IsFocused()

		// Verify no commits were created
		t.Views().Commits().
			IsEmpty()
	},
})