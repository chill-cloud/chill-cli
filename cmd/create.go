package cmd

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cache"
	"github.com/chill-cloud/chill-cli/pkg/integrations/server"
	"github.com/chill-cloud/chill-cli/pkg/service/naming"
	"github.com/go-git/go-git/v5"
	copy2 "github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"html/template"
	"os"
	"path"
	"path/filepath"
)

func RunCreate(cmd *cobra.Command, args []string) error {

	// No project exists for now, so no cwd lookup will be performed
	cwd := Cwd

	name := args[0]
	base := server.DefaultServerName
	if len(args) > 1 {
		base = args[1]
	}

	if !naming.Validate(name) {
		return fmt.Errorf("project name does not follow criteria")
	}

	cacheContext, err := cache.DefaultCacheContext()
	if err != nil {
		return err
	}
	integration := server.ForName(base)
	if integration == nil {
		return fmt.Errorf("no integration found for name %s", base)
	}
	src := cache.GitSource{Remote: integration.GetBaseProjectRemote()}
	err = src.Update(cacheContext)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(cwd, name)
	err = os.MkdirAll(targetPath, os.ModePerm)
	if err != nil {
		return err
	}

	err = copy2.Copy(src.GetPath(cacheContext), targetPath)
	if err != nil {
		return err
	}

	err = os.RemoveAll(filepath.Join(targetPath, ".git"))
	if err != nil {
		return fmt.Errorf("could not remove .git folder: %w\n", err)
	}

	_, err = git.PlainInit(targetPath, false)
	if err != nil {
		return fmt.Errorf("could not init a Git repository: %w\n", err)
	}

	var replacement struct {
		ServiceName     string
		IntegrationName string
	}
	replacement.ServiceName = name
	replacement.IntegrationName = base
	configName := path.Join(targetPath, "chill.yaml")
	tmpl, err := template.New("chill.yaml").ParseFiles(configName)
	if err != nil {
		return fmt.Errorf("could not parse template: %w\n", err)
	}
	out, err := os.OpenFile(configName, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer out.Close()
	err = tmpl.Execute(out, replacement)
	if err != nil {
		return fmt.Errorf("could not process template: %w\n", err)
	}
	return nil
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new Chill service",
	Long:  `Creates a new Chill service`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  RunCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
