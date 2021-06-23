package dbrunner

import (
	"database/sql"
	"embed"
	"fmt"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pkg/errors"
)

// ConnStrProvider returns the full connection string give host and port.
type ConnStrProvider func(host string, port string) string

// DriverInfo provides all the necessary information to start the container.
type DriverInfo struct {
	DriverName string
	Port       string
	RunOpts    dockertest.RunOptions
	ConnStr    ConnStrProvider
}

// New will create a new Runner.
func New(pool *dockertest.Pool) *Runner {
	return &Runner{pool: pool}
}

// Runner holds the necessary resources to create the containers.
type Runner struct {
	pool *dockertest.Pool
}

// DBHandle represents a running DB, it contains the generated connection string and can be used to stop the
// launched database container.
type DBHandle struct {
	ConnStr string
	res     *dockertest.Resource
}

// Close will stop the container and purge the resources from the pool.
func (h *DBHandle) Close() error {
	return h.res.Close()
}

// Run will start a new container based on the provider DriverInfo, wait for the DB to start and
// run the migrations automatically before returning.
func (r *Runner) Run(driver DriverInfo, migrations embed.FS) (*DBHandle, error) {
	res, err := r.pool.RunWithOptions(&driver.RunOpts, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, err
	}

	handler, err := r.connectMigrate(res, driver, migrations)
	if err != nil {
		_ = res.Close()
		return nil, err
	}

	return handler, nil
}

func (r *Runner) connectMigrate(res *dockertest.Resource, driver DriverInfo, migrations embed.FS) (*DBHandle, error) {
	port := res.GetPort(fmt.Sprintf("%s/tcp", driver.Port))
	if port == "" {
		return nil, fmt.Errorf("unable to get port %s", port)
	}

	host := res.GetHostPort(port)
	if host == "" {
		host = "127.0.0.1"
	}

	connStr := driver.ConnStr(host, port)
	r.pool.Retry(func() error {
		db, err := sql.Open(driver.DriverName, connStr)
		if err != nil {
			return err
		}

		defer db.Close()

		return db.Ping()
	})

	migrationDriver, err := httpfs.New(http.FS(migrations), ".")
	if err != nil {
		return nil, errors.Wrap(err, "filed to create migration driver")
	}

	m, err := migrate.NewWithSourceInstance("httpfs", migrationDriver, connStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create source instance")
	}

	err = m.Up()
	if err != nil {
		return nil, errors.Wrap(err, "failed migration up")
	}

	return &DBHandle{
		ConnStr: driver.ConnStr(host, port),
		res:     res,
	}, nil
}
