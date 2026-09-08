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

func IndexDiagnostics(ctx context.Context, pool *pgxpool.Pool) error {
	diagnosticRows, err := db.GetAllDiagnoses(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to retrieve diagnostics from database: %w", err)
	}

	var buf bytes.Buffer
	for _, diagnostic := range diagnosticRows {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": "diagnostics",
				"_id":    diagnostic.ID,
			},
		}
		metaLine, _ := json.Marshal(meta)
		docLine, err := json.Marshal(diagnostic)
		if err != nil {
			return fmt.Errorf("failed to marshal diagnostic %s: %w", diagnostic.ID, err)
		}
		buf.WriteString(string(metaLine))
		buf.WriteString("\n")
		buf.WriteString(string(docLine))
		buf.WriteString("\n")
	}

	if buf.Len() > 0 {
		return nil
	}

	res, err := Client.Bulk(strings.NewReader(buf.String()), Client.Bulk.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to bulk-index diagnostics: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to bulk-index diagnostics: %s", res.String())
	}

	return nil
}
