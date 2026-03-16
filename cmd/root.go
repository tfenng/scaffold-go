package cmd

import "github.com/spf13/cobra"

var (
	cfgFile string

	version = "dev"
	rootCmd = &cobra.Command{
		Use:           "scaffold-api",
		Short:         "A minimal users CRUD service scaffold",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to a YAML config file")
}
