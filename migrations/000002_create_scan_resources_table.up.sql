CREATE TABLE scan_resources (
    id uuid PRIMARY KEY,
    run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    kind text NOT NULL,
    namespace text NOT NULL,
    name text NOT NULL,
    data jsonb NOT NULL
);