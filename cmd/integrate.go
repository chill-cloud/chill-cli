package cmd

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cluster"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/cwd"
	"github.com/chill-cloud/chill-cli/pkg/integrations/remote"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
)

func runIntegrateGithub(cmd *cobra.Command, args []string) error {
	cwd, err := cwd.SetupCwd(Cwd)
	if err != nil {
		return err
	}

	cfg, err := config.ParseConfig(cwd, config.ProjectConfigName, false)
	if err != nil {
		return err
	}
	if cfg == nil {
		return fmt.Errorf("no project config found")
	}

	integration := remote.NewGithub(args[0], args[1], args[2], args[3])
	path, err := cluster.GetKubeconfigPath(Kubeconfig)
	if err != nil {
		return err
	}

	clusterManager, err := cluster.NewForKubernetes(Kubeconfig, KubeNamespace)
	if err != nil {
		return fmt.Errorf("unable to build cluster client")
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Print(err)
	}
	err = integration.SetSecret("CHILL_KUBECONFIG", string(b))

	if err != nil {
		return err
	}

	if !noRegistry {
		err = clusterManager.SetRegistry(cfg.Name, "ghcr.io", args[2], args[3])
		if err != nil {
			return err
		}

		cfg.Registry = fmt.Sprintf("ghcr.io/%s/%s", args[0], args[1])

		newCfg, err := config.ProcessConfig(cfg)
		if err != nil {
			return err
		}
		err = newCfg.SaveToFile(filepath.Join(cwd, config.ProjectConfigName), false)
		if err != nil {
			return err
		}
	}

	return nil
}

var integrateGithub = &cobra.Command{
	Use:   "github <owner_name> <repo_name> <username> <token>",
	Short: "A brief description of your command",
	Args:  cobra.ExactArgs(4),
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: runIntegrateGithub,
}

var integrateCmd = &cobra.Command{
	Use:   "integrate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use one of the subcommands")
	},
}

var noRegistry = false

func init() {
	rootCmd.AddCommand(integrateCmd)
	integrateCmd.Flags().BoolVar(&noRegistry, "no-registry", false, "Disables setting the integration regisrty")
	integrateCmd.AddCommand(integrateGithub)
}
