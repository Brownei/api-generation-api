package tests

import (
	"os"
	"testing"

	"github.com/Brownei/api-generation-api/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadAppConfig_Defaults(t *testing.T) {
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("SERVER_PORT")

	cfg := config.LoadAppConfig()

	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "db", cfg.DBUser)
	assert.Equal(t, "db", cfg.DBPassword)
	assert.Equal(t, "db", cfg.DBName)
	assert.Equal(t, "your-secret-key", cfg.JWTSecret)
	assert.Equal(t, "8080", cfg.ServerPort)
}

func TestLoadAppConfig_FromEnv(t *testing.T) {
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "mydb")
	os.Setenv("JWT_SECRET", "my-secret")
	os.Setenv("SERVER_PORT", "9090")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SERVER_PORT")
	}()

	cfg := config.LoadAppConfig()

	assert.Equal(t, "db.example.com", cfg.DBHost)
	assert.Equal(t, "5433", cfg.DBPort)
	assert.Equal(t, "admin", cfg.DBUser)
	assert.Equal(t, "secret", cfg.DBPassword)
	assert.Equal(t, "mydb", cfg.DBName)
	assert.Equal(t, "my-secret", cfg.JWTSecret)
	assert.Equal(t, "9090", cfg.ServerPort)
}

func TestAppConfig_GetDSN(t *testing.T) {
	cfg := config.LoadAppConfig()
	cfg.DBHost = "localhost"
	cfg.DBPort = "5432"
	cfg.DBUser = "user"
	cfg.DBPassword = "pass"
	cfg.DBName = "testdb"

	dsn := cfg.GetDSN()

	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=user")
	assert.Contains(t, dsn, "password=pass")
	assert.Contains(t, dsn, "dbname=testdb")
}

func TestConnectDB_Invalid(t *testing.T) {
	cfg := config.LoadAppConfig()
	cfg.DBHost = "invalid-host"
	cfg.DBPort = "9999"
	cfg.DBUser = "invalid"
	cfg.DBPassword = "invalid"
	cfg.DBName = "invalid"

	_, err := config.ConnectDB(cfg)

	assert.Error(t, err)
}
