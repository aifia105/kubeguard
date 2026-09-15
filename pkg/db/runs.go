package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Runs struct {
	ID          uuid.UUID `db:"id" json:"id"`
	ClusterID   uuid.UUID `db:"cluster_id" json:"cluster_id"`
	Kind        string    `db:"kind" json:"kind"`
	Scope       string    `db:"scope" json:"scope"`
	StartedAt   time.Time `db:"started_at" json:"started_at"`
	CompletedAt time.Time `db:"completed_at" json:"completed_at"`
	Status      string    `db:"status" json:"status"`
}

func InsertRun(ctx context.Context, pool *pgxpool.Pool, r Runs) (uuid.UUID, error) {
	r.ID = GenerateUUID()
	_, err := pool.Exec(ctx, "INSERT INTO runs (id, cluster_id, kind, scope, status) VALUES ($1, $2, $3, $4, $5)", r.ID, r.ClusterID, r.Kind, r.Scope, r.Status)
	return r.ID, err
}

func GetRunByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (r Runs, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM runs WHERE id=$1", id)
	if err != nil {
		return Runs{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Runs])
}

func GetAllRuns(ctx context.Context, pool *pgxpool.Pool) (runs []Runs, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM runs")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Runs])
}

func GetRunsByClusterID(ctx context.Context, pool *pgxpool.Pool, clusterID uuid.UUID) (runs []Runs, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM runs WHERE cluster_id=$1", clusterID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Runs])
}

func UpdateRunStatus(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, status string) (r Runs, err error) {
	rows, err := pool.Query(ctx, "UPDATE runs SET status=$1 WHERE id=$2 RETURNING *", status, id)
	if err != nil {
		return Runs{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Runs])
}

func UpdateRunCompletedAt(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, completedAt string) (r Runs, err error) {
	rows, err := pool.Query(ctx, "UPDATE runs SET completed_at=$1 WHERE id=$2 RETURNING *", completedAt, id)
	if err != nil {
		return Runs{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Runs])
}
