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
	Use:          "chill-cli",
	Short:        "CLI for Chill backend as a service platform",
	SilenceUsage: true,
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
		defer func() {
			err := logging.Logger.Sync()
			if err != nil {
				panic(err)
			}
		}()
		logging.Logger.Info("Verbose logging enabled")
		return nil
	}
	rootCmd.PersistentFlags().BoolVarP(&v, "verbose", "v", false, "Enable detailed logging")
	rootCmd.PersistentFlags().BoolVarP(&ForceLocal, "local", "l", false, "Force enable local mode")
	rootCmd.PersistentFlags().StringVar(&Cwd, "cwd", "", "Force set project directory")
	rootCmd.PersistentFlags().StringVar(&Kubeconfig, "kubeconfig", "", "Set the kubeconfig path")
	rootCmd.PersistentFlags().StringVar(&KubeNamespace, "kube-namespace", v1.NamespaceDefault, "Set the Kubernetes namespace")
}
