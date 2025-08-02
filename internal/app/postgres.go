package app

import (
	"database/sql"
	"fmt"

	"github.com/guttosm/user-service/config"
	_ "github.com/lib/pq"
)

// InitPostgres initializes a PostgresSQL connection using the provided config.
//
// Parameters:
//   - cfg (*config.Config): The application configuration
//
// Returns:
//   - *sql.DB: The open database connection
//   - error: Any connection or ping failure
func InitPostgres(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
