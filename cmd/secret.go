package cmd

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cluster"
	"github.com/chill-cloud/chill-cli/pkg/service/naming"

	"github.com/spf13/cobra"
)

func runGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	if !naming.Validate(key) {
		return fmt.Errorf("key does not follow criteria")
	}

	clusterManager, err := cluster.NewForKubernetes(Kubeconfig, KubeNamespace)
	if err != nil {
		return fmt.Errorf("unable to build cluster client")
	}

	res, err := clusterManager.GetSecret(key)
	if err != nil {
		return err
	}
	println(res)
	return nil
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	if !naming.Validate(key) {
		return fmt.Errorf("key does not follow criteria")
	}

	value := args[1]
	clusterManager, err := cluster.NewForKubernetes(Kubeconfig, KubeNamespace)
	if err != nil {
		return fmt.Errorf("unable to build cluster client")
	}

	return clusterManager.SetSecret(key, value)
}

func runSecretEnv(cmd *cobra.Command, args []string) error {
	key := args[0]
	if !naming.Validate(key) {
		return fmt.Errorf("key does not follow criteria")
	}
	println(naming.SecretToEnv(key))
	return nil
}

func runSecretMountPath(cmd *cobra.Command, args []string) error {
	key := args[0]
	if !naming.Validate(key) {
		return fmt.Errorf("key does not follow criteria")
	}
	println(naming.SecretToMountPath(key) + "/" + cluster.ChillSecretKey)
	return nil
}

// secretCmd represents the secret command
var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage key-value secrets for Chill services",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use one of the subcommands")
	},
}

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Gets a key-value secret for Chill services",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Sets a key-value secret for Chill services",
	Args:  cobra.ExactArgs(2),
	RunE:  runSet,
}

var secretEnvCmd = &cobra.Command{
	Use:   "env <key>",
	Short: "Prints environment variable name Chill will use to identify this secret in a service",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretEnv,
}

var secretMountPathEnvCmd = &cobra.Command{
	Use:   "mountPath <key>",
	Short: "Prints a mount path that might be read by a Chill service to retrieve a secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretMountPath,
}

func init() {
	rootCmd.AddCommand(secretCmd)

	secretCmd.AddCommand(getCmd)
	secretCmd.AddCommand(setCmd)
	secretCmd.AddCommand(secretEnvCmd)
	secretCmd.AddCommand(secretMountPathEnvCmd)
}
