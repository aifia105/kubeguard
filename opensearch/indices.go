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
	fmt.Printf("DEBUG: index %q exists check returned status %d\n", indexName, exists.StatusCode)

	if exists.StatusCode == 404 {
		res, err := Client.Indices.Create(indexName, Client.Indices.Create.WithBody(strings.NewReader(string(mappings))), Client.Indices.Create.WithContext(ctx))
		if err != nil {
			return err
		}
		fmt.Printf("DEBUG: create %q returned status %d, isError=%v\n", indexName, res.StatusCode, res.IsError())
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
