package cmd

import (
	errors2 "errors"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cache"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/cwd"
	"github.com/spf13/cobra"
	"strings"
)

func RunFreeze(cmd *cobra.Command, args []string) error {
	cwd, err := cwd.SetupCwd(Cwd)
	if err != nil {
		return err
	}

	err = Sync(cwd, true, false)
	if err != nil {
		return err
	}

	cfg, err := config.ParseConfig(cwd, config.LockConfigName, true)
	if err != nil {
		return err
	}

	s, err := cache.NewLocalSourceOfTruth(cwd)
	if err != nil {
		return err
	}

	if cfg.CurrentVersion == nil {
		return fmt.Errorf("no current version set. Use 'chill-cli sync' to initialize the first version")
	}

	err = s.FreezeVersion(*cfg.CurrentVersion)
	if err != nil {
		var typedErr cache.NotCommittedError
		if errors2.As(err, &typedErr) {
			return fmt.Errorf("some changes have not been committed, only committed versions can be frozen;\n"+
				"problematic files:\n\n%s", strings.Join(typedErr.NotCommittedFiles(), "\n"))
		}
		return err
	}

	fmt.Printf("Version %s frozen successfully", cfg.CurrentVersion.String())

	return nil
}

// freezeCmd represents the freeze command
var freezeCmd = &cobra.Command{
	Use:   "freeze",
	Short: "Freezes current version",
	Long: `Freezes current version; your must be synced and
you must not have any untracked and unstaged files
in order to freeze.`,
	RunE: RunFreeze,
}

func init() {
	rootCmd.AddCommand(freezeCmd)
}
