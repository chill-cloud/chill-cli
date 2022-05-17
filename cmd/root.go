package cmd

import (
	"github.com/chill-cloud/chill-cli/pkg/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/core/v1"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chill-cli",
	Short: "CLI for Chill backend as a service platform",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	SilenceUsage: true,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

type BaseProject struct {
	Git string
}

type ChillCLIConfig struct {
	ConfigVersion string
	BaseProjects  map[string]BaseProject
}

var Cwd string
var Kubeconfig string
var KubeNamespace string
var ForceLocal bool

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chill-cli.yaml)")

	var v bool

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		if v {
			config.Level.SetLevel(zap.InfoLevel)
		} else {
			config.Level.SetLevel(zap.WarnLevel)
		}
		config.EncoderConfig.CallerKey = ""
		config.EncoderConfig.TimeKey = ""
		logging.Logger, _ = config.Build()
		defer logging.Logger.Sync() // flushes buffer, if any
		logging.Logger.Info("Verbose logging enabled")
		return nil
	}

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolVarP(&v, "verbose", "v", false, "Enable detailed logging")
	rootCmd.PersistentFlags().BoolVarP(&ForceLocal, "local", "l", false, "Force enable local mode")
	rootCmd.PersistentFlags().StringVar(&Cwd, "cwd", "", "Force set project directory")
	rootCmd.PersistentFlags().StringVar(&Kubeconfig, "kubeconfig", "", "Set the kubeconfig path")
	rootCmd.PersistentFlags().StringVar(&KubeNamespace, "kube-namespace", v1.NamespaceDefault, "Set the Kubernetes namespace")
}
