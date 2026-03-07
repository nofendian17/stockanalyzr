package cmd

import (
	"github.com/spf13/cobra"

	"stockanalyzr/pkg/logger"
	"stockanalyzr/pkg/migration"
	"stockanalyzr/services/user-service/internal/config"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		logger.Info("running migrations up...")
		if err := migration.RunUp(cfg.PostgresDSN, cfg.MigrationsPath); err != nil {
			return err
		}
		logger.Info("migrations applied successfully")
		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback all migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		logger.Info("running migrations down...")
		if err := migration.RunDown(cfg.PostgresDSN, cfg.MigrationsPath); err != nil {
			return err
		}
		logger.Info("migrations rolled back successfully")
		return nil
	},
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd, migrateDownCmd)
	rootCmd.AddCommand(migrateCmd)
}
