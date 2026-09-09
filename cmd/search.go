package cmd

import (
	"encoding/json"

	"github.com/aifia105/kubeguard/opensearch"
	"github.com/aifia105/kubeguard/pkg/audit"
	"github.com/aifia105/kubeguard/pkg/db"
	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	searchIndexFlag     string
	searchSeverityFlag  string
	searchNamespaceFlag string
	searchLimitFlag     int
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

var searchQueryCmd = &cobra.Command{
	Use:   "query [term]",
	Short: "Query the OpenSearch index",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if opensearch.Client == nil {
			logger.LogError("OpenSearch is not available: check KUBEGUARD_OPENSEARCH_URL/USER/PASSWORD")
			return
		}

		var term string
		if len(args) > 0 {
			term = args[0]
		}

		filters := map[string]string{}
		if searchSeverityFlag != "" {
			filters["severity"] = searchSeverityFlag
		}
		if searchNamespaceFlag != "" {
			filters["namespace"] = searchNamespaceFlag
		}

		if term == "" && len(filters) == 0 {
			logger.LogError("provide a search term or at least one filter (--severity, --namespace, --rule-id)")
			return
		}

		results, err := opensearch.ExecuteSearchQuery(ctx, searchIndexFlag, term, filters, searchLimitFlag)
		if err != nil {
			logger.LogError("failed to query OpenSearch: %v", err)
			return
		}

		if len(results) == 0 {
			logger.LogInfo("no results found")
			return
		}

		printSearchResults(results, searchIndexFlag)
	},
}

func printSearchResults(results []json.RawMessage, indexName string) {
	switch indexName {
	case "diagnostics":
		for _, raw := range results {
			var d db.Diagnosis
			if err := json.Unmarshal(raw, &d); err != nil {
				logger.LogWarning("failed to parse result: %v", err)
				continue
			}
			logger.LogInfo("[%s] run:%s model:%s\n%s", d.CreatedAt, d.RunID, d.Model, d.ResponseText)
		}
	case "events":
		for _, raw := range results {
			var e db.Event
			if err := json.Unmarshal(raw, &e); err != nil {
				logger.LogWarning("failed to parse result: %v", err)
				continue
			}
			logger.LogInfo("[%s] %s/%s %s: %s", e.LastSeen, e.Namespace, e.Object, e.Reason, e.Message)
		}
	default:
		var findings []audit.Finding
		for _, raw := range results {
			var f db.Finding
			if err := json.Unmarshal(raw, &f); err != nil {
				logger.LogWarning("failed to parse result: %v", err)
				continue
			}
			findings = append(findings, audit.Finding{
				Namespace:   f.Namespace,
				Name:        f.Name,
				Resource:    f.Resource,
				RuleID:      f.RuleID,
				Severity:    audit.Severity(f.Severity),
				Description: f.Description,
			})
		}
		printFindings(findings, "search results")
	}
}

func init() {
	searchQueryCmd.Flags().StringVar(&searchIndexFlag, "index", "findings", "index to search: findings, diagnostics, events")
	searchQueryCmd.Flags().StringVar(&searchSeverityFlag, "severity", "", "filter by severity (findings only)")
	searchQueryCmd.Flags().StringVar(&searchNamespaceFlag, "namespace", "", "filter by namespace")
	searchQueryCmd.Flags().IntVar(&searchLimitFlag, "limit", 20, "maximum number of results")

	searchCmd.AddCommand(searchSetupCmd, searchIndexCmd, searchQueryCmd)
	rootCmd.AddCommand(searchCmd)
}
