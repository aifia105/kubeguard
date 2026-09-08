package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func CreateIndexIfNotExists(ctx context.Context, indexName string, mappings json.RawMessage) error {
	exists, err := Client.Indices.Exists([]string{indexName}, Client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	if exists.StatusCode == 404 {
		_, err = Client.Indices.Create(indexName, Client.Indices.Create.WithBody(strings.NewReader(string(mappings))), Client.Indices.Create.WithContext(ctx))
		if err != nil {
			return err
		}
	}
	return nil
}

func EnsureAllIndicesExist(ctx context.Context) error {
	if err := CreateIndexIfNotExists(ctx, "findings", json.RawMessage(mappings)); err != nil {
		return fmt.Errorf("failed to create findings index: %w", err)
	}
	if err := CreateIndexIfNotExists(ctx, "diagnostics", json.RawMessage(diagnosticsMappings)); err != nil {
		return fmt.Errorf("failed to create diagnostics index: %w", err)
	}
	if err := CreateIndexIfNotExists(ctx, "events", json.RawMessage(eventsMappings)); err != nil {
		return fmt.Errorf("failed to create events index: %w", err)
	}
	return nil
}
