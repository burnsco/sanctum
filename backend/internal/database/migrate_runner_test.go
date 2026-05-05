package database

import (
	"context"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestValidateAppliedVersions_AllKnown(t *testing.T) {
	registered := []Migration{
		{Version: 1, Name: "baseline"},
		{Version: 2, Name: "posts_poll_fields"},
	}
	applied := []int{1, 2}

	if err := validateAppliedVersions(applied, registered); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateAppliedVersions_UnknownVersion(t *testing.T) {
	registered := []Migration{
		{Version: 1, Name: "baseline"},
		{Version: 2, Name: "posts_poll_fields"},
	}
	applied := []int{1, 3, 2, 99}

	err := validateAppliedVersions(applied, registered)
	if err == nil {
		t.Fatal("expected unknown migration error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "000003") || !strings.Contains(msg, "000099") {
		t.Fatalf("expected unknown versions in error, got %q", msg)
	}
}

func TestApplyMigrationRollsBackSQLWhenLogInsertFails(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&MigrationLog{}); err != nil {
		t.Fatalf("migrate logs: %v", err)
	}
	if err := db.Create(&MigrationLog{Version: 1, Name: "existing"}).Error; err != nil {
		t.Fatalf("seed migration log: %v", err)
	}

	store := NewMigrationStore(db)
	err = store.ApplyMigration(context.Background(), 1, "duplicate_log", "CREATE TABLE should_roll_back (id INTEGER PRIMARY KEY);")
	if err == nil {
		t.Fatal("expected duplicate migration log error")
	}

	if db.Migrator().HasTable("should_roll_back") {
		t.Fatal("expected migration SQL to roll back when log insert fails")
	}
}
