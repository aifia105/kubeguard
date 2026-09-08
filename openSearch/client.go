package opensearch

import (
	"context"
	"fmt"
	"os"

	"github.com/opensearch-project/opensearch-go"
)

var Client *opensearch.Client

func InitOpenSearch(ctx context.Context) error {
	connString := os.Getenv("KUBEGUARD_OPENSEARCH_URL")
	if connString == "" {
		return fmt.Errorf("KUBEGUARD_OPENSEARCH_URL environment variable is not set")
	}
	userName := os.Getenv("KUBEGUARD_OPENSEARCH_USER")
	if userName == "" {
		return fmt.Errorf("KUBEGUARD_OPENSEARCH_USER environment variable is not set")
	}
	passWord := os.Getenv("KUBEGUARD_OPENSEARCH_PASSWORD")
	if passWord == "" {
		return fmt.Errorf("KUBEGUARD_OPENSEARCH_PASSWORD environment variable is not set")
	}

	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{connString},
		Username:  userName,
		Password:  passWord,
	})
	if err != nil {
		return fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	res, err := client.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping OpenSearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("OpenSearch ping returned an error status: %s", res.Status())
	}

	Client = client
	return nil
}
