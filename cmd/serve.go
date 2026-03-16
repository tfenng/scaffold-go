package cmd

import (
	"scaffold-api/internal/config"
	"scaffold-api/pkg/di"

	"github.com/spf13/cobra"
)

func init() {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cmd, cfgFile)
			if err != nil {
				return err
			}

			app := di.NewApp(cfg)
			if err := app.Err(); err != nil {
				return err
			}

			app.Run()
			return nil
		},
	}

	serveCmd.Flags().String("app-name", "", "Application name")
	serveCmd.Flags().String("environment", "", "Runtime environment")
	serveCmd.Flags().String("http-host", "", "HTTP listen host")
	serveCmd.Flags().Int("http-port", 0, "HTTP listen port")
	serveCmd.Flags().Int("http-read-timeout-seconds", 0, "HTTP read timeout in seconds")
	serveCmd.Flags().Int("http-write-timeout-seconds", 0, "HTTP write timeout in seconds")
	serveCmd.Flags().Int("http-shutdown-timeout-seconds", 0, "HTTP shutdown timeout in seconds")
	serveCmd.Flags().String("db-dsn", "", "PostgreSQL DSN")
	serveCmd.Flags().Bool("db-auto-migrate", false, "Auto migrate database schema on startup")
	serveCmd.Flags().String("log-level", "", "Log level: debug, info, warn, error")
	serveCmd.Flags().Bool("log-pretty", false, "Enable console-friendly logs")

	rootCmd.AddCommand(serveCmd)
}
