package database

import (
	"testing"

	"sanctum/internal/config"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConfigurePool(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	cfg := &config.Config{
		DBMaxOpenConns:           10,
		DBMaxIdleConns:           5,
		DBConnMaxLifetimeMinutes: 15,
	}

	err = configurePool(db, cfg)
	assert.NoError(t, err)

	_, err = db.DB()
	assert.NoError(t, err)
}

func TestRedactSQLLiterals(t *testing.T) {
	sql := "INSERT INTO users (email, password) VALUES ('person@example.com', 'hash''withquote') RETURNING id"

	got := redactSQLLiterals(sql)

	assert.NotContains(t, got, "person@example.com")
	assert.NotContains(t, got, "hash")
	assert.Equal(t, "INSERT INTO users (email, password) VALUES (?, ?) RETURNING id", got)
}
