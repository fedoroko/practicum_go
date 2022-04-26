package storage

import (
	"database/sql"
	"github.com/fedoroko/practicum_go/internal/errrs"
	_ "github.com/jackc/pgx/v4/stdlib"
	"sync"

	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/metrics"
)

type postgres struct {
	conn *sql.DB
	cfg  *config.ServerConfig
	mtx  sync.RWMutex
}

type tempMetric struct {
	id    string
	mtype string
	value float64
	delta int64
}

func (t *tempMetric) toMetric() metrics.Metric {
	return metrics.New(
		t.id, t.mtype, t.value, t.delta,
	)
}

func (p *postgres) Get(m metrics.Metric) (metrics.Metric, error) {
	t := tempMetric{}
	getQuery := `SELECT name, type, value, delta
				 FROM metrics
				 WHERE name = $1
				 AND type = $2;`
	err := p.conn.QueryRow(getQuery, m.Name(), m.Type()).
		Scan(&t.id, &t.mtype, &t.value, &t.delta)

	if err != nil {
		return m, err
	}
	ret := t.toMetric()
	if p.cfg.Key != "" {
		ret.SetHash(p.cfg.Key)
	}
	return ret, nil
}

func (p *postgres) Set(m metrics.Metric) error {
	if p.cfg.Key != "" {
		if ok, _ := m.CheckHash(p.cfg.Key); !ok {
			return errrs.ThrowInvalidHashError()
		}
	}
	var exists bool
	checkQuery := `SELECT EXISTS(
						SELECT 1 FROM metrics
						WHERE name = $1
						AND type = $2
					);`
	err := p.conn.QueryRow(checkQuery, m.Name(), m.Type()).
		Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		updateQuery := `UPDATE metrics
						SET value = $1, delta = delta + $2
						WHERE name = $3
						AND type = $4;`
		if _, err = p.conn.Exec(
			updateQuery,
			m.Float64Value(),
			m.Int64Value(),
			m.Name(),
			m.Type(),
		); err != nil {
			return err
		}
	} else {
		insertQuery := `INSERT INTO metrics (name, type, value, delta)
						VALUES($1, $2, $3, $4);`
		if _, err = p.conn.Exec(
			insertQuery,
			m.Name(),
			m.Type(),
			m.Float64Value(),
			m.Int64Value(),
		); err != nil {
			return err
		}
	}

	return nil
}

func (p *postgres) List() ([]metrics.Metric, error) {
	var ret []metrics.Metric

	getQuery := `SELECT name, type, value, delta 
				 FROM metrics
				 ORDER BY type DESC, name ASC;`
	rows, err := p.conn.Query(getQuery)
	if err != nil {
		return ret, err
	}

	for rows.Next() {
		t := tempMetric{}
		if err = rows.Scan(&t.id, &t.mtype, &t.value, &t.delta); err != nil {
			return ret, err
		}

		ret = append(ret, t.toMetric())
	}

	if err = rows.Err(); err != nil {
		return ret, err
	}

	return ret, nil
}

func (p *postgres) Ping() error {
	return p.conn.Ping()
}

func (p *postgres) Close() error {
	return p.conn.Close()
}

func postgresInterface(cfg *config.ServerConfig) *postgres {
	conn, err := sql.Open("pgx", cfg.Database)
	if err != nil {
		panic(err)
	}

	var exists bool
	existsQuery := `SELECT EXISTS (
						SELECT FROM pg_tables
						WHERE schemaname = 'public'
						AND tablename = 'metrics'
    				);`
	err = conn.QueryRow(existsQuery).Scan(&exists)
	if err != nil || !exists {
		createQuery := `CREATE TABLE metrics (
							id serial PRIMARY KEY,
							name VARCHAR (50) NOT NULL,
							type VARCHAR (20) NOT NULL,
							value DOUBLE PRECISION,
							delta INT,
							updated_at TIMESTAMP
						);`
		_, err = conn.Exec(createQuery)

		if err != nil {
			panic(err)
		}
	}

	return &postgres{
		conn: conn,
		cfg:  cfg,
		mtx:  sync.RWMutex{},
	}
}
