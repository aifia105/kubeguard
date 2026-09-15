package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aifia105/kubeguard/pkg/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func IndexFinding(ctx context.Context, pool *pgxpool.Pool) error {
	findingRows, err := db.GetAllFindings(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to retrieve findings from database: %w", err)
	}

	var buf bytes.Buffer
	for _, finding := range findingRows {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": "findings",
				"_id":    finding.ID,
			},
		}
		metaLine, _ := json.Marshal(meta)
		docLine, err := json.Marshal(finding)
		if err != nil {
			return fmt.Errorf("failed to marshal finding %s: %w", finding.ID, err)
		}
		buf.WriteString(string(metaLine))
		buf.WriteString("\n")
		buf.WriteString(string(docLine))
		buf.WriteString("\n")
	}

	if buf.Len() == 0 {
		return nil
	}

	res, err := Client.Bulk(strings.NewReader(buf.String()), Client.Bulk.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to bulk-index findings: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("bulk index request failed: %s", res.Status())
	}
	return nil
}
