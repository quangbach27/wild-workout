//go:build integration

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
	testutils.RunMigrations("user", embedMigrations, "migrations")
	os.Exit(m.Run())
}
