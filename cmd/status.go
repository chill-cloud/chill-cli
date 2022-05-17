package cmd

import (
	"context"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cluster"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/util"
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

	cwd, err := util.SetupCwd(Cwd)
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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: RunStatus,
}

var Major int

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().IntVarP(&Major, "major", "m", 0, "Overrides major version defined in the lock file")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
