package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

func queryBuilder(term string, filters map[string]string, index string) (json.RawMessage, error) {
	var clauses []interface{}

	if index != "" {
		if term != "" {
			switch index {
			case "findings":
				clauses = append(clauses, map[string]interface{}{
					"match": map[string]interface{}{
						"description": term,
					},
				})
			case "diagnostics":
				clauses = append(clauses, map[string]interface{}{
					"match": map[string]interface{}{
						"response_text": term,
					},
				})
			case "events":
				clauses = append(clauses, map[string]interface{}{
					"match": map[string]interface{}{
						"message": term,
					},
				})
			default:
				clauses = append(clauses, map[string]interface{}{
					"match": map[string]interface{}{
						"description": term,
					},
				})
			}
		}
	}

	var filterClauses []interface{}
	for field, value := range filters {
		filterClauses = append(filterClauses, map[string]interface{}{
			"term": map[string]interface{}{
				field: value,
			},
		})
	}

	boolQuery := map[string]interface{}{}

	if len(clauses) > 0 {
		boolQuery["must"] = clauses
	}

	if len(filterClauses) > 0 {
		boolQuery["filter"] = filterClauses
	}

	return json.Marshal(map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
	})
}

func ExecuteSearchQuery(ctx context.Context, index string, term string, filters map[string]string, limit int) ([]json.RawMessage, error) {
	query, err := queryBuilder(term, filters, index)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %v", err)
	}

	res, err := Client.Search(
		Client.Search.WithIndex(index),
		Client.Search.WithBody(bytes.NewReader(query)),
		Client.Search.WithContext(ctx),
		Client.Search.WithPretty(),
		Client.Search.WithSize(limit),
	)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("opensearch returned an error: %s", res.String())
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source json.RawMessage `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %v", err)
	}

	sources := make([]json.RawMessage, 0, len(result.Hits.Hits))

	for _, hit := range result.Hits.Hits {
		sources = append(sources, hit.Source)
	}
	return sources, nil
}
