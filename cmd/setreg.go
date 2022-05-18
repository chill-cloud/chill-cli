/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cluster"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/cwd"
	"github.com/spf13/cobra"
	"path/filepath"
)

func RunSetReg(cmd *cobra.Command, args []string) error {
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

	server := args[0]
	login := args[1]
	password := args[2]

	clusterManager, err := cluster.NewForKubernetes(Kubeconfig, KubeNamespace)
	if err != nil {
		return fmt.Errorf("unable to build cluster client")
	}

	err = clusterManager.SetRegistry(cfg.Name, server, login, password)
	if err != nil {
		return err
	}

	reg := server
	if infix != "" {
		reg += "/" + infix
	}
	cfg.Registry = reg

	newCfg, err := config.ProcessConfig(cfg)
	if err != nil {
		return err
	}
	err = newCfg.SaveToFile(filepath.Join(cwd, config.ProjectConfigName), false)
	if err != nil {
		return err
	}

	q, w, err := clusterManager.GetRegistry(cfg.Name, server)
	if err != nil {
		return err
	}

	println(q, w)

	return nil
}

var infix string

// setregCmd represents the setreg command
var setregCmd = &cobra.Command{
	Use:   "setreg <server> <login> <password>",
	Args:  cobra.MinimumNArgs(3),
	Short: "Sets container registry for this service",
	Long: `Sets container registry for this service, and also
saves its credentials into the cluster`,
	RunE: RunSetReg,
}

func init() {
	rootCmd.AddCommand(setregCmd)

	setregCmd.Flags().StringVarP(&infix, "infix", "i", "", "Image infix")
}
