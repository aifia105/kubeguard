package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Diagnosis struct {
	ID               uuid.UUID       `db:"id"`
	RunID            uuid.UUID       `db:"run_id"`
	Model            string          `db:"model"`
	ResponseText     string          `db:"response_text"`
	EvidenceSnapshot json.RawMessage `db:"evidence_snapshot"`
	CreatedAt        time.Time       `db:"created_at"`
}

func InsertDiagnosis(ctx context.Context, pool *pgxpool.Pool, d Diagnosis) (uuid.UUID, error) {
	d.ID = GenerateUUID()
	_, err := pool.Exec(ctx, "INSERT INTO diagnoses (id, run_id, model, response_text, evidence_snapshot) VALUES ($1, $2, $3, $4, $5)", d.ID, d.RunID, d.Model, d.ResponseText, d.EvidenceSnapshot)
	return d.ID, err
}

func GetDiagnosisByRunID(ctx context.Context, pool *pgxpool.Pool, runID uuid.UUID) (d Diagnosis, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM diagnoses WHERE run_id=$1", runID)
	if err != nil {
		return Diagnosis{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Diagnosis])
}

func GetAllDiagnoses(ctx context.Context, pool *pgxpool.Pool) (diagnoses []Diagnosis, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM diagnoses")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Diagnosis])
}

func GetDiagnosisByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (d Diagnosis, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM diagnoses WHERE id=$1", id)
	if err != nil {
		return Diagnosis{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Diagnosis])
}
