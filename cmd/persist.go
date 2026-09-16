package cmd

import (
	"encoding/json"

	"github.com/aifia105/kubeguard/pkg/audit"
	"github.com/aifia105/kubeguard/pkg/db"
	"github.com/aifia105/kubeguard/pkg/evidence"
	"github.com/aifia105/kubeguard/pkg/logger"
	"github.com/google/uuid"
)

func toDBFinding(runID uuid.UUID, f audit.Finding) db.Finding {
	return db.Finding{
		ID:          db.GenerateUUID(),
		RunID:       runID,
		Namespace:   f.Namespace,
		Name:        f.Name,
		Resource:    f.Resource,
		RuleID:      f.RuleID,
		Severity:    string(f.Severity),
		Description: f.Description,
	}
}

func toDBEvent(runID uuid.UUID, e evidence.Event) db.Event {
	return db.Event{
		ID:        db.GenerateUUID(),
		RunID:     runID,
		Namespace: e.Namespace,
		Object:    e.Object,
		Type:      e.Type,
		Reason:    e.Reason,
		Message:   e.Message,
		Count:     e.Count,
		LastSeen:  e.LastSeen,
	}
}

func persistRun(runID uuid.UUID, kind, scope string, findings []audit.Finding) {
	if noSaveFlag || db.Pool == nil || clusterID == uuid.Nil {
		return
	}

	run := db.Runs{ID: runID, ClusterID: clusterID, Kind: kind, Scope: scope, Status: "success"}
	runID, err := db.InsertRun(ctx, db.Pool, run)
	if err != nil {
		if db.IsSchemaNotExist(err) {
			logger.LogWarning("database schema not initialized; results will not be saved. Run: kubeguard db migrate")
			return
		}
		logger.LogWarning("failed to save run to database: %v", err)
		return
	}
	for _, f := range findings {
		dbFinding := toDBFinding(runID, f)
		if _, err := db.InsertFinding(ctx, db.Pool, dbFinding); err != nil {
			logger.LogWarning("failed to save finding: %v", err)
		}
	}
}

func persistEvents(runID uuid.UUID, events []evidence.Event) {
	if noSaveFlag || db.Pool == nil {
		return
	}
	for _, e := range events {
		dbEvent := toDBEvent(runID, e)
		if _, err := db.InsertEvent(ctx, db.Pool, dbEvent); err != nil {
			logger.LogWarning("failed to save event: %v", err)
		}
	}
}

func persistScanResource(runID uuid.UUID, kind, namespace, name string, data []byte) {
	if noSaveFlag || db.Pool == nil {
		return
	}
	row := db.Scan_resources{ID: db.GenerateUUID(), RunID: runID, Kind: kind, Namespace: namespace, Name: name, Data: data}
	if _, err := db.InsertScanResource(ctx, db.Pool, row); err != nil {
		logger.LogWarning("failed to save scan resource: %v", err)
	}
}

func persistScan(scope string, results scanResults) {
	if noSaveFlag || db.Pool == nil || clusterID == uuid.Nil {
		return
	}
	run := db.Runs{ID: db.GenerateUUID(), ClusterID: clusterID, Kind: "scan", Scope: scope, Status: "success"}
	runID, err := db.InsertRun(ctx, db.Pool, run)
	if err != nil {
		if db.IsSchemaNotExist(err) {
			logger.LogWarning("database schema not initialized — run: kubeguard db migrate")
			return
		}
		logger.LogWarning("failed to save run: %v", err)
		return
	}

	for _, pod := range results.Pods {
		data, _ := json.Marshal(pod)
		persistScanResource(runID, "Pod", pod.Namespace, pod.Name, data)
	}
	redactedSecrets := redactSecrets(results.Secrets)
	for i, secret := range results.Secrets {
		data, _ := json.Marshal(redactedSecrets[i])
		persistScanResource(runID, "Secret", secret.Namespace, secret.Name, data)
	}

	for _, configMap := range results.ConfigMaps {
		data, _ := json.Marshal(configMap)
		persistScanResource(runID, "ConfigMap", configMap.Namespace, configMap.Name, data)
	}

	for _, service := range results.Services {
		data, _ := json.Marshal(service)
		persistScanResource(runID, "Service", service.Namespace, service.Name, data)
	}

	for _, deployment := range results.Deployments {
		data, _ := json.Marshal(deployment)
		persistScanResource(runID, "Deployment", deployment.Namespace, deployment.Name, data)
	}

	for _, event := range results.Events {
		data, _ := json.Marshal(event)
		persistScanResource(runID, "Event", event.Namespace, event.Name, data)
	}

	for _, ingress := range results.Ingresses {
		data, _ := json.Marshal(ingress)
		persistScanResource(runID, "Ingress", ingress.Namespace, ingress.Name, data)
	}

	for _, limitRange := range results.LimitRanges {
		data, _ := json.Marshal(limitRange)
		persistScanResource(runID, "LimitRange", limitRange.Namespace, limitRange.Name, data)
	}
	for _, resourceQuota := range results.ResourceQuotas {
		data, _ := json.Marshal(resourceQuota)
		persistScanResource(runID, "ResourceQuota", resourceQuota.Namespace, resourceQuota.Name, data)
	}
	for _, namespace := range results.Namespaces {
		data, _ := json.Marshal(namespace)
		persistScanResource(runID, "Namespace", namespace.Name, namespace.Name, data)
	}

	for _, node := range results.Nodes {
		data, _ := json.Marshal(node)
		persistScanResource(runID, "Node", "", node.Name, data)
	}

	for _, networkPolicy := range results.NetworkPolicies {
		data, _ := json.Marshal(networkPolicy)
		persistScanResource(runID, "NetworkPolicy", networkPolicy.Namespace, networkPolicy.Name, data)
	}

	for _, nodeMetrics := range results.NodesMetrics {
		data, _ := json.Marshal(nodeMetrics)
		persistScanResource(runID, "NodeMetrics", "", nodeMetrics.Name, data)
	}

	for _, podMetrics := range results.PodsMetrics {
		data, _ := json.Marshal(podMetrics)
		persistScanResource(runID, "PodMetrics", podMetrics.Namespace, podMetrics.Name, data)
	}

}
