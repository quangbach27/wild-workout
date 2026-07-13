package db_test

import (
	"embed"
	"os"
	"testing"
	"workout/common/testutils"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func TestMain(m *testing.M) {
	testutils.RunMigrations("trainer", embedMigrations, "migrations")
	os.Exit(m.Run())
}
