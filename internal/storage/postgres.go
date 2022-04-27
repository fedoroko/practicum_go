package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/errrs"
	"github.com/fedoroko/practicum_go/internal/metrics"
	_ "github.com/jackc/pgx/v4/stdlib"
	"io"
	"strings"
)

type postgres struct {
	*sql.DB
	upsertStmt *sql.Stmt
	getStmt    *sql.Stmt
	listStmt   *sql.Stmt
	buffer     []metrics.Metric
	cfg        *config.ServerConfig
}

type tempMetric struct {
	Id    string   `json:"id"`
	Type  string   `json:"type"`
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
}

const (
	schema string = `CREATE TABLE metrics (
							id serial PRIMARY KEY,
							name VARCHAR (50) UNIQUE NOT NULL,
							type VARCHAR (20) NOT NULL,
							value DOUBLE PRECISION,
							delta BIGINT,
							updated_at TIMESTAMP
						);`

	getQuery string = `SELECT name, type, value, delta
					   FROM metrics
					   WHERE name = $1
					   AND type = $2;`

	listQuery string = `SELECT name, type, value, delta 
						FROM metrics
						ORDER BY type DESC, name ASC;`

	upsertQuery string = `INSERT INTO metrics (name, type, value, delta)
						  VALUES($1, $2, $3, $4)
						  ON CONFLICT(name)	DO UPDATE
 						  SET value = $3, delta = metrics.delta + $4`
)

func (t *tempMetric) toMetric() metrics.Metric {
	name := t.Id
	s := strings.Split(t.Id, "::")
	if len(s) > 1 {
		name = s[0]
	}
	return metrics.NewOmitEmpty(
		name, t.Type, t.Value, t.Delta,
	)
}

func (p *postgres) Get(m metrics.Metric) (metrics.Metric, error) {
	t := tempMetric{}
	err := p.getStmt.QueryRow(m.Name(), m.Type()).
		Scan(&t.Id, &t.Type, &t.Value, &t.Delta)

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

	name := m.Name() + "::" + m.Type()
	_, err := p.upsertStmt.Exec(
		name,
		m.Type(),
		m.Float64Pointer(),
		m.Int64Pointer(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgres) SetBatch(j io.Reader) error {
	if p.DB == nil {
		return errors.New("no db")
	}

	decoder := json.NewDecoder(j)
	for {
		t := tempMetric{}
		if err := decoder.Decode(&t); errors.Is(err, io.EOF) {
			return p.flush()
		} else if err != nil {
			return err
		}

		m := t.toMetric()
		if err := p.addMetric(m); err != nil {
			return err
		}
	}
}

func (p *postgres) List() ([]metrics.Metric, error) {
	var ret []metrics.Metric

	rows, err := p.listStmt.Query()
	if err != nil {
		return ret, err
	}

	for rows.Next() {
		t := tempMetric{}
		if err = rows.Scan(&t.Id, &t.Type, &t.Value, &t.Delta); err != nil {
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
	return p.DB.Ping()
}

func (p *postgres) Close() error {
	return p.DB.Close()
}

func (p *postgres) addMetric(m metrics.Metric) error {
	p.buffer = append(p.buffer, m)
	if cap(p.buffer) == len(p.buffer) {
		if err := p.flush(); err != nil {
			return err
		}
	}

	return nil
}

func (p *postgres) flush() error {
	if p.DB == nil {
		return errors.New("no db")
	}

	tx, err := p.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(upsertQuery)
	if err != nil {
		return err
	}

	for _, m := range p.buffer {
		name := m.Name() + "::" + m.Type()
		if _, err = stmt.Exec(name, m.Type(), m.Float64Pointer(), m.Int64Pointer()); err != nil {
			if err = tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	p.buffer = p.buffer[:0]
	return nil
}

func postgresInterface(cfg *config.ServerConfig) *postgres {
	db, err := sql.Open("pgx", cfg.Database)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			panic(err)
		}
	}

	getStmt, err := db.Prepare(getQuery)
	if err != nil {
		panic(err)
	}

	listStmt, err := db.Prepare(listQuery)
	if err != nil {
		panic(err)
	}

	upsertStmt, err := db.Prepare(upsertQuery)
	if err != nil {
		panic(err)
	}

	return &postgres{
		DB:         db,
		getStmt:    getStmt,
		listStmt:   listStmt,
		upsertStmt: upsertStmt,
		cfg:        cfg,
		buffer:     make([]metrics.Metric, 0, 100),
	}
}
