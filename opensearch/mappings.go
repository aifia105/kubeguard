package opensearch

var mappings = ` {
    "properties": {
      "id": { "type": "keyword" },
      "run_id": { "type": "keyword" },
      "cluster_id": { "type": "keyword" },
      "namespace": { "type": "keyword" },
      "name": { "type": "keyword" },
      "resource": { "type": "keyword" },
      "rule_id": { "type": "keyword" },
      "severity": { "type": "keyword" },
      "description": { "type": "text" },
      "started_at": { "type": "date" },
	    "cluster_name": { "type": "keyword" }
    }
}`

var diagnosticsMappings = `{
    "properties": {
      "id": { "type": "keyword" },
      "run_id": { "type": "keyword" },
      "cluster_id": { "type": "keyword" },
      "model": { "type": "keyword" },
      "response_text": { "type": "text" },
      "created_at": { "type": "date" }
    }
}`

var eventsMappings = `{
	"properties": {
	  "id": { "type": "keyword" },
	  "run_id": { "type": "keyword" },
	  "namespace": { "type": "keyword" },
	  "object": { "type": "keyword" },
	  "type": { "type": "keyword" },
	  "reason": { "type": "keyword" },
	  "message": { "type": "text" },
	  "count": { "type": "integer" },
	  "last_seen": { "type": "date" }
	}
}`
