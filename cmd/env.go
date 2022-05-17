package cmd

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/service/naming"
	"github.com/chill-cloud/chill-cli/pkg/util"
	"github.com/spf13/cobra"
)

func RunEnv(cmd *cobra.Command, args []string) error {
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		cwd, err := util.SetupCwd(Cwd)
		if err != nil {
			return err
		}
		cfg, err := config.ParseConfig(cwd, config.LockConfigName, true)
		if err != nil {
			return err
		}
		if cfg == nil {
			return fmt.Errorf("no project config found")
		}
		name = cfg.Name
	}
	println(naming.NameToEnv(name))
	return nil
}

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: RunEnv,
}

func init() {
	rootCmd.AddCommand(envCmd)
}
