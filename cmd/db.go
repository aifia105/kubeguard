package cmd

import (
	"log"

	"github.com/aifia105/kubeguard/pkg/db"
	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/spf13/cobra"
)

var migrateDownSteps int

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage database schema migrations",
}

var migrateUpCmd = &cobra.Command{
	Use:   "migrate-up",
	Short: "Apply all pending migrations to the database",
	Run: func(cmd *cobra.Command, args []string) {
		if err := db.MigrateUp(); err != nil {
			log.Fatalf("Failed to migrate up: %v", err)
		}
		logger.LogSuccess("database schema is up to date")
	},
}

var dbMigrateDownCmd = &cobra.Command{
	Use:   "migrate-down",
	Short: "Roll back the most recent migration(s)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := db.MigrateDown(migrateDownSteps); err != nil {
			logger.LogFatal("rollback failed: %v", err)
			return
		}
		logger.LogSuccess("rolled back %d migration(s)", migrateDownSteps)
	},
}

var dbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current migration status",
	Run: func(cmd *cobra.Command, args []string) {
		version, dirty, err := db.MigrationStatus()
		if err != nil {
			logger.LogFatal("failed to get migration status: %v", err)
			return
		}
		if dirty {
			logger.LogWarning("schema is at version %d but marked DIRTY — a previous migration failed partway. Investigate before running further migrations.", version)
			return
		}
		logger.LogSuccess("schema is at version %d", version)
	},
}

func init() {
	dbMigrateDownCmd.Flags().IntVar(&migrateDownSteps, "steps", 1, "number of migrations to roll back")
	dbCmd.AddCommand(migrateUpCmd, dbMigrateDownCmd, dbStatusCmd)
	rootCmd.AddCommand(dbCmd)
}
