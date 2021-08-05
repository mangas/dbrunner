package dbrunner

import (
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest/v3"
)

const (
	PostgresDefaultUser     = "postgres"
	PostgresDefaultPassword = "postgres"
	MySQLDefaultUser        = "mysql"
	MySQLDefaultPassword    = "mysql"
)

func NewMySQLDriverInfo(username string, password string) DriverInfo {
	return DriverInfo{
		DriverName: "mysql",
		Port:       "3306",
		RunOpts: dockertest.RunOptions{
			Repository: "mysql/mysql-server",
			Tag:        "latest",
			Env: []string{
				// "MYSQL_ROOT_PASSWORD=mysql",
				// "MYSQL_ROOT_HOST=%",
				fmt.Sprintf("MYSQL_USER=%s", username),
				fmt.Sprintf("MYSQL_PASSWORD=%s", password),
				"MYSQL_DATABASE=bananas",
			},
		},
		ConnStr: func(host, port string) string {
			c := mysql.Config{
				User:                 username,
				Passwd:               password,
				Net:                  "tcp",
				Addr:                 fmt.Sprintf("%s:%s", host, port),
				TLSConfig:            "skip-verify",
				DBName:               "bananas",
				AllowNativePasswords: true,
			}

			return fmt.Sprintf("mysql://%s", c.FormatDSN())
		},
	}
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

var DefaultMysqlDriverInfo DriverInfo = NewMySQLDriverInfo(MySQLDefaultUser, MySQLDefaultPassword)
var DefaultPostgresDriverInfo DriverInfo = NewPostgresDriverInfo(PostgresDefaultUser, PostgresDefaultPassword)
