CREATE TABLE events (
    id uuid PRIMARY KEY,
    run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    namespace text NOT NULL,
    object text NOT NULL,
    type text NOT NULL,
    reason text NOT NULL,
    message text NOT NULL,
    count integer NOT NULL,
    last_seen_at timestamptz NOT NULL DEFAULT now()
);