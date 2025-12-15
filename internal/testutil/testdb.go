package testutil

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/infra/repository"
)

type TestDB struct {
	Container testcontainers.Container
	DB        *gorm.DB
	DSN       string
}

func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := runMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return &TestDB{
		Container: pgContainer,
		DB:        db,
		DSN:       dsn,
	}
}

func (tdb *TestDB) TeardownTestDB(t *testing.T) {
	t.Helper()

	if err := tdb.Container.Terminate(context.Background()); err != nil {
		t.Logf("failed to terminate container: %v", err)
	}
}

func (tdb *TestDB) CleanTable(t *testing.T) {
	t.Helper()

	if err := tdb.DB.Exec("TRUNCATE TABLE reminds").Error; err != nil {
		t.Fatalf("failed to clean table: %v", err)
	}
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&repository.RemindModel{})
}
