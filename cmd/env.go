package cmd

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/cwd"
	"github.com/chill-cloud/chill-cli/pkg/service/naming"
	"github.com/spf13/cobra"
)

func RunEnv(cmd *cobra.Command, args []string) error {
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		cwd, err := cwd.SetupCwd(Cwd)
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
	Short: "Prints environment variable name with a dependency address",
	Long: `This command prints environment variable name
with a dependency address to be used inside the service code.
`,
	RunE: RunEnv,
}

func init() {
	rootCmd.AddCommand(envCmd)
}
