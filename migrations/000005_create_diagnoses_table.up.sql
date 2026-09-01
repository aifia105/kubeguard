CREATE TABLE diagnoses (
    id uuid PRIMARY KEY,
    run_id uuid NOT NULL UNIQUE REFERENCES runs(id) ON DELETE CASCADE,
    model text NOT NULL,
    response_text text NOT NULL,
    evidence_snapshot jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);