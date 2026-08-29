package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Scan_resources struct {
	ID        uuid.UUID       `db:"id"`
	RunID     uuid.UUID       `db:"run_id"`
	Kind      string          `db:"kind"`
	Name      string          `db:"name"`
	Namespace string          `db:"namespace"`
	Data      json.RawMessage `db:"data"`
}

func InsertScanResource(ctx context.Context, pool *pgxpool.Pool, s Scan_resources) (uuid.UUID, error) {
	s.ID = GenerateUUID()
	_, err := pool.Exec(ctx, "INSERT INTO scan_resources (id, run_id, kind, name, namespace, data) VALUES ($1, $2, $3, $4, $5, $6)", s.ID, s.RunID, s.Kind, s.Name, s.Namespace, s.Data)
	return s.ID, err
}

func GetScanResourceByID(ctx context.Context, pool *pgxpool.Pool, id string) (s Scan_resources, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM scan_resources WHERE id=$1", id)
	if err != nil {
		return Scan_resources{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Scan_resources])
}

func GetAllScanResources(ctx context.Context, pool *pgxpool.Pool) (resources []Scan_resources, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM scan_resources")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Scan_resources])
}

func GetScanResourceByName(ctx context.Context, pool *pgxpool.Pool, name string) (s Scan_resources, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM scan_resources WHERE name=$1", name)
	if err != nil {
		return Scan_resources{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Scan_resources])
}

func GetScanResourcesByRunID(ctx context.Context, pool *pgxpool.Pool, runID string) (resources []Scan_resources, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM scan_resources WHERE run_id=$1", runID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Scan_resources])
}
