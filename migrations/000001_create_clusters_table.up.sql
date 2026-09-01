 CREATE TABLE clusters (
    id uuid PRIMARY KEY,
    name text NOT NULL UNIQUE,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);


