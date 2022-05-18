package cmd

import (
	"context"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cluster"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/cwd"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	url2 "net/url"
	"os"
	"strconv"
	"strings"
)

func RunStatus(cmd *cobra.Command, args []string) error {
	clusterManager, err := cluster.NewForKubernetes(Kubeconfig, KubeNamespace)
	if err != nil {
		return fmt.Errorf("unable to build cluster client")
	}
	knative, err := clusterManager.GetKnative()
	if err != nil {
		return fmt.Errorf("unable to build Knative client")
	}

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

	var major int
	if Major == 0 {
		major = cfg.CurrentVersion.GetMajor()
	} else {
		major = Major
	}

	fmt.Printf("Service %s, major version %d\n\n", cfg.Name, major)

	res, err := knative.Services(KubeNamespace).Get(
		context.TODO(),
		clusterManager.GetServiceIdentifier(cfg.Name, *cfg.CurrentVersion),
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("unable to fetch the service")
	}
	mainUrl, err := url2.Parse(res.Status.URL.String())
	if err != nil {
		return fmt.Errorf("unable to parse URL: %w\n", err)
	}

	fmt.Printf("Main host: %s\n\n", mainUrl.Host)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Version", "Host", "Traffic percent"})

	for _, q := range res.Status.Traffic {
		url, err := url2.Parse(q.URL.String())
		if err != nil {
			return fmt.Errorf("unable to parse URL: %w\n", err)
		}
		firstPart := strings.Split(url.Host, ".")[0]
		_, version, err := clusterManager.GetServiceAndVersion(firstPart)
		if err != nil {
			return fmt.Errorf("could not get version from the host")
		}
		table.Append([]string{
			version.String(),
			url.Host,
			strconv.FormatInt(*q.Percent, 10),
		})
	}
	table.Render()
	return nil
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Prints status of the current major deployment",
	RunE:  RunStatus,
}

var Major int

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().IntVarP(&Major, "major", "m", 0, "Overrides major version defined in the lock file")
}
