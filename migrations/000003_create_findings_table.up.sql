CREATE TABLE findings (
    id uuid PRIMARY KEY,
    run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    namespace text NOT NULL,
    name text NOT NULL,
    resource text NOT NULL,
    rule_id text NOT NULL,
    severity text  NOT NULL CHECK (severity IN ('critical', 'high', 'medium', 'low')),
    description text
)

CREATE INDEX idx_findings_run_id ON findings(run_id);

CREATE INDEX idx_findings_severity ON findings(severity);