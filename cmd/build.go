package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/cwd"
	"github.com/chill-cloud/chill-cli/pkg/logging"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"io"
	"strings"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds an image of the service",
	Long: `This command connects to the Docker daemon and tries to build
your image declared in image/Dockerfile, then, if successful, 
marks it with a tag of the current version.`,
	RunE: RunBuild,
}

func GetContext(filePath string) (io.Reader, error) {
	_, err := homedir.Expand(filePath)
	if err != nil {
		return nil, err
	}
	ctx, err := archive.TarWithOptions(filePath, &archive.TarOptions{})
	return ctx, err
}

func RunBuild(cmd *cobra.Command, args []string) error {

	cwd, err := cwd.SetupCwd(Cwd)
	if err != nil {
		return err
	}

	cfg, err := config.ParseConfig(cwd, config.LockConfigName, true)
	if err != nil {
		return err
	}

	logging.Logger.Info("Creating Docker client...")
	cli, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return fmt.Errorf("unable to connect to the Docker daemon: %w\n", err)
	}
	ctx := context.Background()
	dockerCtx, err := GetContext(cwd)
	if err != nil {
		return fmt.Errorf("unable to create Docker context: %w\n", err)
	}
	imageName, _ := cfg.GetBuildTag(ForceLocal)
	logging.Logger.Info(fmt.Sprintf("Building an image %s...", imageName))
	resp, err := cli.ImageBuild(ctx, dockerCtx, types.ImageBuildOptions{
		Dockerfile: "image/Dockerfile",
		Tags:       []string{imageName},
	})
	if err != nil {
		return fmt.Errorf("unable to build image: %w\n", err)
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
		}
		if jm.Error != nil {
			return fmt.Errorf("unable to build image: %w\n", jm.Error)
		} else {
			var out strings.Builder
			jm.Display(&out, false)
			res := out.String()
			to := len(res) - 1
			if to < 0 {
				to = 0
			}
			logging.Logger.Info(fmt.Sprintf("[Docker] %s", res[:to]))
		}
	}
	fmt.Printf("Image has been built successfully with tag %s\n", imageName)
	return nil
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
