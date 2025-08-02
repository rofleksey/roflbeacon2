package migration

import (
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/elliotchance/pie/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do"
	"log/slog"
	"roflbeacon2/pkg/database"
	"time"
)

type Migration interface {
	Id() string
	Execute(ctx context.Context, di *do.Injector, tx pgx.Tx, queries *database.Queries) error
}

var allMigrations = []Migration{}

func doExecute(
	ctx context.Context,
	di *do.Injector,
	dbConn *pgxpool.Pool,
	queries *database.Queries,
	migration Migration,
) error {
	tx, err := dbConn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := queries.WithTx(tx)

	if err = migration.Execute(ctx, di, tx, qtx); err != nil {
		return fmt.Errorf("error executing migration body: %w", err)
	}

	if _, err = queries.CreateMigration(ctx, database.CreateMigrationParams{
		ID:      migration.Id(),
		Applied: time.Now(),
	}); err != nil {
		return fmt.Errorf("error inserting migration: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func Migrate(ctx context.Context, di *do.Injector) error {
	slog.LogAttrs(ctx, slog.LevelInfo, "Executing migrations...")

	dbConn := do.MustInvoke[*pgxpool.Pool](di)
	queries := do.MustInvoke[*database.Queries](di)

	executedMigrations, err := queries.GetMigrations(ctx)
	if err != nil {
		return fmt.Errorf("could not get migrations: %w", err)
	}

	executedIds := mapset.NewThreadUnsafeSet[string]()
	for _, migration := range executedMigrations {
		executedIds.Add(migration.ID)
	}

	pendingMigrations := pie.Filter(allMigrations, func(m Migration) bool {
		return !executedIds.Contains(m.Id())
	})

	if len(pendingMigrations) == 0 {
		slog.LogAttrs(ctx, slog.LevelInfo, "No pending migrations")
		return nil
	}

	for _, migration := range pendingMigrations {
		slog.LogAttrs(ctx, slog.LevelInfo, "Starting migration",
			slog.String("id", migration.Id()),
		)

		if err = doExecute(ctx, di, dbConn, queries, migration); err != nil {
			return fmt.Errorf("could not execute migration %v: %w", migration.Id(), err)
		}

		slog.LogAttrs(ctx, slog.LevelInfo, "Migration success",
			slog.String("id", migration.Id()),
		)
	}

	log.Info("Migrations complete")

	return nil
}
