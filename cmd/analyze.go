package cmd

import (
	"context"
	"os"

	"github.com/aifia105/kubeguard/ollama"
	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	analyzeModel string
	analyzePath  string
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
}
