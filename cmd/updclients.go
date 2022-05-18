package cmd

import (
	"fmt"
	cache2 "github.com/chill-cloud/chill-cli/pkg/cache"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/cwd"
	"github.com/chill-cloud/chill-cli/pkg/integrations/client"
	"github.com/spf13/cobra"
)

func RunUpdclients(cmd *cobra.Command, args []string) error {
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

	cacheContext, err := cache2.DefaultCacheContext()
	if err != nil {
		return err
	}

	for t, remote := range cfg.Clients {
		c := client.ForName(t)

		src := cache2.GitSource{Remote: remote}
		err := src.Update(cacheContext)
		if err != nil {
			return err
		}
		if c == nil {
			return fmt.Errorf("client integration not found for name %s", t)
		}

		err = client.GenerateClient(c, cwd, src.GetPath(cacheContext), cfg.Name, *cfg.CurrentVersion)
		if err != nil {
			return err
		}
	}
	return nil
}

// updclientsCmd represents the updclients command
var updclientsCmd = &cobra.Command{
	Use:   "updclients",
	Short: "Updates code generation for external clients",
	RunE:  RunUpdclients,
}

func init() {
	rootCmd.AddCommand(updclientsCmd)
}
