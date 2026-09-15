package db

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Finding struct {
	ID          uuid.UUID `db:"id" json:"id"`
	RunID       uuid.UUID `db:"run_id" json:"run_id"`
	Namespace   string    `db:"namespace" json:"namespace"`
	Name        string    `db:"name" json:"name"`
	Resource    string    `db:"resource" json:"resource"`
	RuleID      string    `db:"rule_id" json:"rule_id"`
	Severity    string    `db:"severity" json:"severity"`
	Description string    `db:"description" json:"description"`
}

func InsertFinding(ctx context.Context, pool *pgxpool.Pool, f Finding) (uuid.UUID, error) {
	f.ID = GenerateUUID()
	f.Severity = strings.ToLower(f.Severity)
	_, err := pool.Exec(ctx, "INSERT INTO findings (id, run_id, namespace, name, resource, rule_id, severity, description) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", f.ID, f.RunID, f.Namespace, f.Name, f.Resource, f.RuleID, f.Severity, f.Description)
	return f.ID, err
}
func GetAllFindings(ctx context.Context, pool *pgxpool.Pool) (findings []Finding, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM findings")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Finding])
}

func GetFindingByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (f Finding, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM findings WHERE id=$1", id)
	if err != nil {
		return Finding{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Finding])
}

func GetFindingsByRunID(ctx context.Context, pool *pgxpool.Pool, runID uuid.UUID) (findings []Finding, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM findings WHERE run_id=$1", runID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Finding])
}
