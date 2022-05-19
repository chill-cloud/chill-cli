package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cluster"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/cwd"
	"github.com/chill-cloud/chill-cli/pkg/logging"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"io"
	"net/url"
	"os/exec"
	"strings"
)

func RunPush(cmd *cobra.Command, args []string) error {
	cwd, err := cwd.SetupCwd(Cwd)
	if err != nil {
		return err
	}

	logging.Logger.Info("Creating Docker client...")
	cli, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return fmt.Errorf("unable to connect to the Docker daemon: %w\n", err)
	}

	cfg, err := config.ParseConfig(cwd, config.LockConfigName, true)
	if err != nil {
		return err
	}
	imageName, isLocal := cfg.GetBuildTag(ForceLocal)

	fmt.Printf("Pushing image %s...\n", imageName)

	if isLocal {
		q := exec.Command(
			"minikube",
			"-p",
			minikubeProfile,
			"image",
			"load",
			imageName,
		)
		o, err := q.Output()
		if err != nil {
			return err
		}
		logging.Logger.Info(string(o))
	} else {
		username := "oauth2accesstoken"
		if token == "" {
			logging.Logger.Info("No token set, trying to query from Kubernetes...")

			clusterManager, err := cluster.NewForKubernetes(Kubeconfig, KubeNamespace)
			if err != nil {
				return fmt.Errorf("unable to build cluster client: %w", err)
			}
			var host string
			regUrl, err := url.Parse(cfg.Registry)
			if err != nil {
				return err
			}
			if regUrl.Host == "" {
				host = strings.Split(cfg.Registry, "/")[0]
			} else {
				host = regUrl.Host
			}

			username, token, err = clusterManager.GetRegistry(cfg.Name, host)
			if err != nil {
				return err
			}
		}
		authConfig := types.AuthConfig{
			Username: username,
			Password: token,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return fmt.Errorf("error when encoding authConfig. err: %w", err)
		}

		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		pushResp, err := cli.ImagePush(context.TODO(), imageName, types.ImagePushOptions{
			All:          true,
			RegistryAuth: authStr,
		})
		if err != nil {
			return fmt.Errorf("Unable to push image: %w\n", err)
		}
		var out strings.Builder
		_, err = io.Copy(&out, pushResp)
		if err != nil {
			return err
		}
		println(out.String())
	}
	fmt.Printf("Image pushed successfully!")
	return nil
}

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: RunPush,
}

var minikubeProfile string
var token string

func init() {
	rootCmd.AddCommand(pushCmd)

	deployCmd.Flags().StringVar(
		&minikubeProfile,
		"minikube-profile",
		"knative",
		"Minikube profile where Knative is installed to")

	deployCmd.Flags().StringVarP(
		&token,
		"token",
		"t",
		"",
		"Registry OAuth token (will be queried from Kubernetes if not set)")
}
