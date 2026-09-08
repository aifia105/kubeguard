package cmd

import (
	opensearch "github.com/aifia105/kubeguard/openSearch"
	"github.com/aifia105/kubeguard/pkg/db"
	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Manage and query the OpenSearch index",
}

var searchSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create the OpenSearch indices (findings, diagnostics, events) if they don't exist",
	Run: func(cmd *cobra.Command, args []string) {
		if opensearch.Client == nil {
			logger.LogError("OpenSearch is not available: check KUBEGUARD_OPENSEARCH_URL/USER/PASSWORD")
			return
		}
		if err := opensearch.EnsureAllIndicesExist(ctx); err != nil {
			logger.LogError("failed to set up indices: %v", err)
			return
		}
		logger.LogSuccess("OpenSearch indices are ready")
	},
}

var searchIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Sync findings, diagnostics, and events from Postgres into OpenSearch",
	Run: func(cmd *cobra.Command, args []string) {
		if opensearch.Client == nil {
			logger.LogError("OpenSearch is not available: check KUBEGUARD_OPENSEARCH_URL/USER/PASSWORD")
			return
		}
		if db.Pool == nil {
			logger.LogError("database is not available: check KUBEGUARD_POSTGRES_URL")
			return
		}

		if err := opensearch.IndexFinding(ctx, db.Pool); err != nil {
			logger.LogError("failed to index findings: %v", err)
		} else {
			logger.LogSuccess("findings indexed")
		}

		if err := opensearch.IndexDiagnostics(ctx, db.Pool); err != nil {
			logger.LogError("failed to index diagnostics: %v", err)
		} else {
			logger.LogSuccess("diagnostics indexed")
		}

		if err := opensearch.IndexEvents(ctx, db.Pool); err != nil {
			logger.LogError("failed to index events: %v", err)
		} else {
			logger.LogSuccess("events indexed")
		}
	},
}

func init() {
	searchCmd.AddCommand(searchSetupCmd, searchIndexCmd)
	rootCmd.AddCommand(searchCmd)
}
