package cmd

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cache"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/util"
	"github.com/spf13/cobra"
	"strings"
)

func RunFreeze(cmd *cobra.Command, args []string) error {
	cwd, err := util.SetupCwd(Cwd)
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
		if notCommited, ok := err.(cache.NotCommitedError); ok {
			return fmt.Errorf("some changes have not been commited, only commited versions can be frozen;\n"+
				"problematic files:\n\n%s", strings.Join(notCommited.NotCommitedFiles(), "\n"))
		}
		return err
	}

	fmt.Printf("Version %s frozen successfully", cfg.CurrentVersion.String())

	return nil
}

// freezeCmd represents the freeze command
var freezeCmd = &cobra.Command{
	Use:   "freeze",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: RunFreeze,
}

func init() {
	rootCmd.AddCommand(freezeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// freezeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// freezeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
