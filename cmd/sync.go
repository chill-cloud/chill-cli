package cmd

import (
	errors2 "errors"
	"fmt"
	cache2 "github.com/chill-cloud/chill-cli/pkg/cache"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/cwd"
	"github.com/chill-cloud/chill-cli/pkg/integrations/server"
	"github.com/chill-cloud/chill-cli/pkg/logging"
	"github.com/chill-cloud/chill-cli/pkg/service"
	"github.com/chill-cloud/chill-cli/pkg/validate"
	"github.com/chill-cloud/chill-cli/pkg/version"
	"github.com/chill-cloud/chill-cli/pkg/version/constraint"
	"github.com/chill-cloud/chill-cli/pkg/version/set"
	"github.com/spf13/cobra"
	"os/exec"
	"path"
	"path/filepath"
)

func Sync(cwd string, local bool, gen bool) error {
	local = local || ForceLocal

	// Set up cache

	cacheContext, err := cache2.DefaultCacheContext()
	if err != nil {
		return err
	}

	// Parse project config

	cfg, err := config.ParseConfig(cwd, config.ProjectConfigName, false)
	if err != nil {
		return err
	}
	if cfg == nil {
		return fmt.Errorf("no project config found")
	}

	// Parse lock file (if exists)

	lockCfg, err := config.ParseConfig(cwd, config.LockConfigName, true)
	if err != nil {
		return err
	}
	if lockCfg == nil {
		lockCfg = new(service.ProjectConfig)
	}

	// Copy changes from the config to the lock file

	if err := cfg.ApplyIdempotent(lockCfg); err != nil {
		return err
	}

	// Set actual versions of dependencies

	for d := range cfg.Dependencies {
		logging.Logger.Info(fmt.Sprintf("Updating dependency %s", d.GetName()))
		if !local {
			err := d.Cache().Update(cacheContext)
			if err != nil {
				return err
			}
		}
		vs, err := d.Cache().GetVersions(cacheContext)
		if err != nil {
			return err
		}
		q := set.ArrayVersionSet(vs)
		err = q.Validate()
		if err != nil {
			return err
		}

		// We can freely depend on development-grade APIs because they are
		// proven to be compatible (on protocol level at least) with
		// its root production version (vx.y.0)

		v := q.GetLatestVersion(d.GetVersion())

		if v == nil {
			return fmt.Errorf("no version matching constraints %s", d.GetVersion())
		}

		logging.Logger.Info(fmt.Sprintf("Best matching version is %s", v.String()))

		err = d.Cache().SwitchToVersion(cacheContext, *v)
		if err != nil {
			return fmt.Errorf("%s: unable to switch to version %s: %w", d.GetName(), v.String(), err)
		}
		err = d.SetSpecificVersion(v)
		if err != nil {
			return fmt.Errorf("unable to set specific version %s: %w", v.String(), err)
		}
		logging.Logger.Info(fmt.Sprintf("Switched to %s", v.String()))
	}

	err = validate.ValidateGraph(cfg, cacheContext, local)

	if err != nil {
		return fmt.Errorf("invalid service specification: %w\n", err)
	}

	sourceOfTruth, err := cache2.NewLocalSourceOfTruth(cwd)

	if err != nil {
		return err
	}

	// Checking if we want to change version

	vs, err := sourceOfTruth.GetVersions()
	versionSet := set.ArrayVersionSet(vs)
	if err != nil {
		return fmt.Errorf("unable to query versions")
	}

	if len(versionSet) == 0 {
		// First version, we should just save lock file
		cfg.CurrentVersion = &version.Version{Major: 1, Minor: 0, Patch: 0}
		cfg.Stage = service.StageMajor
	} else {
		if cfg.BaseVersion == nil {
			// The version after the first one
			cfg.BaseVersion = cfg.CurrentVersion
		}
		// Not the first version, we should check for API breaking changes
		if cfg.Stage == service.StageProduction {
			cfg.BaseVersion = versionSet.GetLatestVersion(constraint.NewMinorOnly(constraint.New(
				version.Version{Major: cfg.BaseVersion.GetMajor(), Minor: 0, Patch: 0},
				version.Version{Major: cfg.BaseVersion.GetMajor() + 1, Minor: 0, Patch: 0},
			)))
			cfg.CurrentVersion = version.New(cfg.BaseVersion.GetMajor(), cfg.BaseVersion.GetMinor()+1, 0)
		}
		lookupTag := fmt.Sprintf("chill-%s", cfg.BaseVersion.String())
		protoPath := path.Join(cwd, "api")
		found, err := sourceOfTruth.CheckVersion(*cfg.BaseVersion)
		logging.Logger.Info(fmt.Sprintf("Looking up for a tag for version %s", cfg.BaseVersion.String()))
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("tried to get a tag of version %s but could not find it: %w", cfg.BaseVersion.String(), err)
		}
		logging.Logger.Info("Tag found; looking for breaking changes...")
		q := exec.Command(
			"buf",
			"breaking",
			protoPath,
			"--against",
			fmt.Sprintf(filepath.Join(cwd, ".git")+"#tag=%s,subdir=api", lookupTag),
		)
		_, err = q.Output()
		if err != nil {
			var exitErr *exec.ExitError
			if errors2.As(err, &exitErr); exitErr.ExitCode() == 100 {
				println("Breaking changes detected! Incrementing major version")
				cfg.Stage = service.StageMajor
			} else {
				return err
			}
		} else {
			logging.Logger.Info("No breaking changes found")
		}
		switch cfg.Stage {
		case service.StageProduction:
			// We already set all the needed changes
			break
		case service.StageMajor:
			v := versionSet.GetLatestVersion(constraint.Any())
			cfg.CurrentVersion = version.New(v.GetMajor()+1, 0, 0)
			if !cfg.BaseVersion.MayBeNext(cfg.CurrentVersion) {
				println(
					"WARNING! It looks like somebody increased a major version before you did.\n" +
						"It is an unusual situation, and is a sign that development process is out of sync.\n" +
						"We suggest to get into touch with latest changes of the project before writing the code.")
			}
		case service.StageDevelopment:
			v := versionSet.GetLatestVersion(constraint.New(*cfg.BaseVersion,
				version.Version{
					Major: cfg.BaseVersion.GetMajor(),
					Minor: cfg.BaseVersion.GetMajor() + 1,
					Patch: 0,
				},
			))
			cfg.CurrentVersion = version.New(v.GetMajor(), v.GetMinor(), v.GetPatch()+1)
		}
	}

	// Save lock file

	newLock, err := config.ProcessConfig(cfg)
	if err != nil {
		return err
	}
	err = newLock.SaveToFile(filepath.Join(cwd, config.LockConfigName), true)
	if err != nil {
		return err
	}

	// generate new apis

	logging.Logger.Info("Generating server stubs for this project...")
	integration := server.ForName(cfg.Integration)
	srcCwd := filepath.Join(cwd, "src")
	if integration == nil {
		return fmt.Errorf("no integration found for name %s", cfg.Integration)
	}
	err = integration.GenerateMethods(srcCwd, cfg.Name, cwd)
	if err != nil {
		return err
	}

	if gen {
		for d := range cfg.Dependencies {
			err = integration.GenerateMethods(srcCwd, d.GetName(), d.Cache().GetPath(cacheContext))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func RunSync(cmd *cobra.Command, args []string) error {
	// Set up working directory
	cwd, err := cwd.SetupCwd(Cwd)
	if err != nil {
		return err
	}
	return Sync(cwd, true, true)
}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronizes local and remote state",
	Long: `You are free to run this command as often as you want
until you are ready to freeze`,
	RunE: RunSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
