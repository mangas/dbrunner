package dbrunner

import (
	"fmt"

	"github.com/ory/dockertest/v3"
)

const (
	PostgresDefaultUser     = "postgres"
	PostgresDefaultPassword = "postgres"
)

func NewMySQLDriverInfo(username string, password string) DriverInfo {
	return DriverInfo{}
}

func NewPostgresDriverInfo(username string, password string) DriverInfo {
	return DriverInfo{
		DriverName: "pgx",
		Port:       "5432",
		RunOpts: dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "13-alpine",
			Env: []string{
				fmt.Sprintf("POSTGRES_USER=%s", username),
				fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
				"listen_addresses='*'",
			},
		},
		ConnStr: func(host, port string) string {
			return fmt.Sprintf("postgres://%s:%s@%s:%s?sslmode=disable", username, password, host, port)
		},
	}
}

var DefaultMysqlDriverInfo DriverInfo = DriverInfo{}
var DefaultPostgresDriverInfo DriverInfo = NewPostgresDriverInfo(PostgresDefaultUser, PostgresDefaultPassword)
