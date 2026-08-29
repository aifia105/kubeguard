package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Event struct {
	ID        uuid.UUID `db:"id"`
	RunID     uuid.UUID `db:"run_id"`
	Namespace string    `db:"namespace"`
	Object    string    `db:"object"`
	Type      string    `db:"type"`
	Reason    string    `db:"reason"`
	Message   string    `db:"message"`
	Count     int32     `db:"count"`
	LastSeen  string    `db:"last_seen"`
}

func InsertEvent(ctx context.Context, pool *pgxpool.Pool, e Event) (uuid.UUID, error) {
	e.ID = GenerateUUID()
	_, err := pool.Exec(ctx, "INSERT INTO events (id, run_id, namespace, object, type, reason, message, count) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", e.ID, e.RunID, e.Namespace, e.Object, e.Type, e.Reason, e.Message, e.Count)
	return e.ID, err
}

func GetAllEvents(ctx context.Context, pool *pgxpool.Pool) (events []Event, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM events")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Event])
}

func GetEventByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (e Event, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM events WHERE id=$1", id)
	if err != nil {
		return Event{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Event])
}

func GetEventsByRunID(ctx context.Context, pool *pgxpool.Pool, runID uuid.UUID) (events []Event, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM events WHERE run_id=$1", runID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Event])
}

func UpdateEventCountAndLastSeen(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, count int32, lastSeen string) error {
	_, err := pool.Exec(ctx, "UPDATE events SET count=$1, last_seen=$2 WHERE id=$3", count, lastSeen, id)
	return err
}
