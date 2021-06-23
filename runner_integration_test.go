// +build integration_test

package dbrunner_test

import (
	"database/sql"
	"testing"

	"github.com/mangas/dbrunner"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/mangas/dbrunner/testdata/migration"
)

func TestRunnerPostgres(t *testing.T) {
	suite.Run(t, &TestSuite{driverInfo: dbrunner.DefaultPostgresDriverInfo})
}

type TestSuite struct {
	suite.Suite

	driverInfo dbrunner.DriverInfo
	runner     *dbrunner.Runner
}

func (s *TestSuite) SetupSuite() {
	pool, err := dockertest.NewPool("")
	s.Require().NoError(err)

	s.runner = dbrunner.New(pool)
}

func (s *TestSuite) TestMigration() {
	require := s.Require()

	handle, err := s.runner.Run(s.driverInfo, migration.Files)
	require.NoError(err)

	s.T().Cleanup(func() {
		handle.Close()
	})

	db, err := sql.Open(s.driverInfo.DriverName, handle.ConnStr)
	require.NoError(err)

	insert(s.T(), 100, db)
	is := query(s.T(), db)
	require.ElementsMatch([]int{100}, is)
}

func insert(t *testing.T, n int, db *sql.DB) {
	res, err := db.Exec("INSERT INTO example (id) VALUES ($1)", n)
	require.NoError(t, err)

	rs, err := res.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rs)
}

func query(t *testing.T, db *sql.DB) []int {
	var is []int
	rows, err := db.Query("SELECT * FROM example")
	require.NoError(t, err)

	var i int
	for rows.Next() {
		err = rows.Scan(&i)
		require.NoError(t, err)

		is = append(is, i)
	}

	return is
}
