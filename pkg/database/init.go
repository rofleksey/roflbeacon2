package database

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

func InitSchema(ctx context.Context, dbConn *pgxpool.Pool) error {
	_, err := dbConn.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("sql exec error: %w", err)
	}

	return nil
}
