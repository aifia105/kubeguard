package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Cluster struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func InsertCluster(ctx context.Context, pool *pgxpool.Pool, c Cluster) (uuid.UUID, error) {
	c.ID = GenerateUUID()
	_, err := pool.Exec(ctx, "INSERT INTO clusters (id, name) VALUES ($1, $2)", c.ID, c.Name)
	return c.ID, err
}

func GetClusterByName(ctx context.Context, pool *pgxpool.Pool, name string) (c Cluster, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM clusters WHERE name=$1", name)
	if err != nil {
		return Cluster{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Cluster])
}

func GetAllClusters(ctx context.Context, pool *pgxpool.Pool) (clusters []Cluster, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM clusters")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Cluster])
}

func GetClusterByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (c Cluster, err error) {
	rows, err := pool.Query(ctx, "SELECT * FROM clusters WHERE id=$1", id)
	if err != nil {
		return Cluster{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Cluster])
}

func GetOrCreateCluster(ctx context.Context, pool *pgxpool.Pool, name string) (c Cluster, err error) {
	c, err = GetClusterByName(ctx, pool, name)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Cluster{}, err
	}
	newCluster := Cluster{ID: GenerateUUID(), Name: name}
	if _, err := InsertCluster(ctx, pool, newCluster); err != nil {
		return Cluster{}, err
	}
	return newCluster, nil
}
