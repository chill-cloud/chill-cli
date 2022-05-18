package cmd

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cache"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/cwd"
	service2 "github.com/chill-cloud/chill-cli/pkg/service"
	"github.com/chill-cloud/chill-cli/pkg/version/constraint"
	"github.com/chill-cloud/chill-cli/pkg/version/set"
	"github.com/spf13/cobra"
	"path/filepath"
)

func RunAddRemote(cmd *cobra.Command, args []string) error {
	var forceVersion *string
	if len(args) > 1 {
		forceVersion = &args[1]
	}
	return runAddGeneric(
		&cache.GitSource{Remote: args[0]},
		forceVersion,
		func(s string, v constraint.Constraint) service2.Dependency {
			return &service2.RemoteDependency{
				Name:    s,
				Git:     args[0],
				Version: v,
			}
		})
}

func RunAddLocal(cmd *cobra.Command, args []string) error {
	var forceVersion *string
	if len(args) > 1 {
		forceVersion = &args[1]
	}
	return runAddGeneric(
		&cache.GitLocalSource{LocalPath: args[0]},
		forceVersion,
		func(s string, v constraint.Constraint) service2.Dependency {
			return &service2.LocalDependency{
				Name:    s,
				Path:    args[0],
				Version: v,
			}
		})
}

func runAddGeneric(src cache.CachedSource, version *string, f func(string, constraint.Constraint) service2.Dependency) error {
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

	cacheContext, err := cache.DefaultCacheContext()
	if err != nil {
		return err
	}
	if !ForceLocal {
		err = src.Update(cacheContext)
		if err != nil {
			return fmt.Errorf("unable to get dependency")
		}
	}

	depCfg, err := config.ParseConfig(src.GetPath(cacheContext), config.LockConfigName, true)
	if err != nil {
		return fmt.Errorf("unable to parse a lock file of the dependency")
	}

	for dep, _ := range cfg.Dependencies {
		if dep.GetName() == depCfg.Name {
			return fmt.Errorf("dependency already present")
		}
	}

	var c constraint.Constraint

	versions, err := src.GetVersions(cacheContext)
	if err != nil {
		return fmt.Errorf("unable to get version list")
	}

	verset := set.ArrayVersionSet(versions)

	if version != nil {
		c, err = constraint.ParseFromString(*version)
		if err != nil {
			return err
		}
		if verset.GetLatestVersion(c) == nil {
			return fmt.Errorf("no supported versions present")
		}
	} else {
		if len(verset) == 0 {
			return fmt.Errorf("no versions present")
		}
		c, err = constraint.ParseFromString(fmt.Sprintf("v%d", verset.GetLatestMajorVersion()))
		if err != nil {
			return err
		}
	}

	cfg.Dependencies[f(depCfg.Name, c)] = true

	newCfg, err := config.ProcessConfig(cfg)
	if err != nil {
		return err
	}
	err = newCfg.SaveToFile(filepath.Join(cwd, config.ProjectConfigName), false)
	if err != nil {
		return err
	}

	return nil
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds another service as a dependency",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use one of the subcommands")
	},
}

var addRemoteCmd = &cobra.Command{
	Use:   "remote <remote_path>",
	Short: "Adds remote dependency",
	RunE:  RunAddRemote,
}

var addLocalCmd = &cobra.Command{
	Use:   "local <local_path>",
	Short: "Adds local dependency",
	RunE:  RunAddLocal,
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.AddCommand(addLocalCmd)
	addCmd.AddCommand(addRemoteCmd)
}
