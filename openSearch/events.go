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

func IndexEvents(ctx context.Context, pool *pgxpool.Pool) error {
	eventRows, err := db.GetAllEvents(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to retrieve events from database: %w", err)
	}

	var buf bytes.Buffer
	for _, event := range eventRows {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": "events",
				"_id":    event.ID,
			},
		}
		metaLine, _ := json.Marshal(meta)
		docLine, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event %s: %w", event.ID, err)
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
		return fmt.Errorf("failed to bulk-index events: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("bulk index request failed: %s", res.Status())
	}
	return nil

}
