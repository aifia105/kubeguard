CREATE TABLE runs (
    id uuid PRIMARY KEY,
    cluster_id uuid NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('scan', 'audit', 'diagnose', 'analyze')),
    scope text NOT NULL,
    started_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    status text NOT NULL CHECK (status IN ('partial_failure', 'success', 'failed')),
    updated_at timestamptz NOT NULL DEFAULT now()
)

CREATE INDEX idx_runs_cluster_id ON runs(cluster_id, kind, started_at);