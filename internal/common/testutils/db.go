package testutils

import (
	"context"
	"io/fs"
	"os"
	"workout/common"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(moduleName string, embedFS fs.FS, migrationsDir string) {
	ctx := context.Background()
	pool, cleanUp := NewDB(ctx)
	defer cleanUp()

	if err := common.MigrateDatabaseUp(ctx, moduleName, pool, embedFS, migrationsDir); err != nil {
		panic(err)
	}
}

func NewDB(ctx context.Context) (*pgxpool.Pool, func()) {
	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5432/wild-workout?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}

	cleanUpFn := func() {
		pool.Close()
	}

	return pool, cleanUpFn
}
