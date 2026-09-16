package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aifia105/kubeguard/ollama"
	"github.com/aifia105/kubeguard/pkg/db"
	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	analyzeModel    string
	analyzePath     string
	analyzeSaveToDB bool
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [namespace] [model] [evidence_bundle.json]",
	Short: "Send a gathered evidence bundle to Ollama for AI-assisted diagnosis",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runAnalyze()
	},
}

func init() {
	analyzeCmd.Flags().StringVar(&analyzeModel, "model", "phi3", "Ollama model to use")
	analyzeCmd.Flags().StringVar(&analyzePath, "file", "output/diagnose_results.json", "path to the evidence bundle JSON file")
	analyzeCmd.Flags().BoolVar(&analyzeSaveToDB, "save", false, "save the diagnosis result to the database")
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze() {
	evidence, err := os.ReadFile(analyzePath)
	if err != nil {
		logger.LogError("no evidence bundle found at %s — run `kubeguard diagnose` first", analyzePath)
		return
	}

	logger.LogInfo("Sending evidence bundle to Ollama (model: %s)...", analyzeModel)
	client := ollama.NewClient(analyzeModel)
	result, err := client.Chat(context.Background(), evidence)
	if err != nil {
		logger.LogError("failed to send evidence bundle to Ollama: %v", err)
		return
	}

	logger.LogInfo("Analysis result: %s", result)

	if !analyzeSaveToDB {
		return
	}
	if db.Pool == nil {
		logger.LogWarning("database unavailable — diagnosis not saved")
		return
	}

	var snap struct {
		RunID uuid.UUID `json:"runId"`
	}
	if err := json.Unmarshal(evidence, &snap); err != nil || snap.RunID == uuid.Nil {
		logger.LogWarning("evidence bundle has no run_id — diagnosis not saved. Regenerate it with a current `kubeguard diagnose`")
		return
	}

	diagnosis := db.Diagnosis{
		ID:               db.GenerateUUID(),
		RunID:            snap.RunID,
		Model:            analyzeModel,
		ResponseText:     result,
		EvidenceSnapshot: evidence,
	}
	_, err = db.InsertDiagnosis(ctx, db.Pool, diagnosis)
	if err != nil {
		if db.IsSchemaNotExist(err) {
			logger.LogWarning("database schema not initialized; results will not be saved. Run: kubeguard db migrate")
			return
		}
		logger.LogWarning("failed to save run to database: %v", err)
		return
	}
	persistDiagnosis(snap.RunID, diagnosis)

}
