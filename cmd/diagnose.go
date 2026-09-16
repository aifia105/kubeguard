package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/aifia105/kubeguard/pkg/collectors"
	"github.com/aifia105/kubeguard/pkg/db"
	"github.com/aifia105/kubeguard/pkg/evidence"
	"github.com/aifia105/kubeguard/pkg/jsonoutput"
	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/spf13/cobra"
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Gather cluster evidence (pod status, events, logs, audit findings) for diagnosis",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ns := resolveNamespace(args)
		RunDiagnose(ns)
	},
}

func init() {
	rootCmd.AddCommand(diagnoseCmd)
}

func RunDiagnose(namespace string) {
	logger.LogInfo("Running full audit for evidence...")
	findings := CollectFullAuditFindings(namespace)

	runID := db.GenerateUUID()
	persistRun(runID, "diagnose", namespace, findings)

	logger.LogInfo("Gathering pod status, container status, and events...")
	pods, err := collectors.ListPods(ctx, k8sClient, namespace)
	if err != nil {
		logger.LogError("Error occurred while listing pods: %v", err)
	}
	events, err := collectors.ListEvents(ctx, k8sClient, namespace)
	if err != nil {
		logger.LogError("Error occurred while listing events: %v", err)
	}
	warningEvents := evidence.FilterWarningsEvents(events)
	persistEvents(runID, warningEvents)
	logger.LogInfo("Fetching recent logs for unhealthy pods...")
	cluster, err := collectors.ClusterInfo(k8sClient)
	clusterName := ""
	if err != nil {
		logger.LogError("Error occurred while fetching cluster info: %v", err)
	} else {
		clusterName = cluster.GitVersion
	}

	snapshot := evidence.BuildSnapshot(ctx, k8sClient, mclientset, runID, namespace, clusterName, findings, pods, events)

	if outputFlag == "json" {
		path := "output/diagnose_results.json"
		if err := jsonoutput.Write(path, snapshot); err != nil {
			logger.LogFatal("failed to write JSON output: %v", err)
			return
		}
		logger.LogInfo("wrote evidence bundle to %s", path)
		return
	}

	logger.LogSuccess("Evidence bundle ready: %d finding(s), %d pod(s) of interest, %d warning event(s), %d log excerpt(s)",
		snapshot.Summary.Total, len(snapshot.Pods), len(snapshot.Events), len(snapshot.Logs))

	evidenceJSON, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		logger.LogError("failed to marshal evidence snapshot: %v", err)
		return
	}
	fmt.Println(string(evidenceJSON))
}
