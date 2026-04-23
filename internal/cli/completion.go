package cli

// `tene completion <shell>` — emits a shell completion script to stdout.
// Cobra ships the generators; we only wire the command so it is discoverable
// and so GoReleaser / Homebrew can package the output.
//
// Usage examples:
//
//   # Bash (one-off)
//   source <(tene completion bash)
//
//   # Zsh
//   tene completion zsh > "${fpath[1]}/_tene"
//
//   # Fish
//   tene completion fish > ~/.config/fish/completions/tene.fish
//
//   # PowerShell
//   tene completion powershell | Out-String | Invoke-Expression

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate a shell completion script for tene.

Examples:

  # Bash (one-off)
  source <(tene completion bash)

  # Bash (persistent, via Homebrew)
  tene completion bash > $(brew --prefix)/etc/bash_completion.d/tene

  # Zsh
  tene completion zsh > "${fpath[1]}/_tene"

  # Fish
  tene completion fish > ~/.config/fish/completions/tene.fish

  # PowerShell
  tene completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			_ = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			_ = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			_ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}
