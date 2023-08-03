/* MIT License

Copyright (c) 2023 Asaduzzaman Pavel (contact@iampavel.dev)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package datastore

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	createSqliteTableQuery = `
		CREATE TABLE IF NOT EXISTS job_postings (
			platform TEXT,
			id TEXT,
			url TEXT,
			job_title TEXT,
			company TEXT,
			applied INTEGER,
			PRIMARY KEY (platform, id)
		);

		CREATE TABLE IF NOT EXISTS applied_counts (
			platform TEXT,
			date TEXT,
			count INTEGER,
			PRIMARY KEY (platform, date)
		);

		CREATE TABLE IF NOT EXISTS applied_counts_by_company (
			name TEXT,
			date TEXT,
			count INTEGER,
			PRIMARY KEY (name, date)
		);
	`
)

var _ Datastore = (*sqlite)(nil)

type sqlite struct {
	db *sql.DB
}

// GetAppliedCountByCompany implements Datastore.
func (d *sqlite) GetAppliedCountByCompany(ctx context.Context, name string) (int, error) {
	row := d.db.QueryRowContext(ctx, `SELECT count FROM applied_counts_by_company WHERE name = ?`, name)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return 0, nil
		}

		return 0, row.Err()
	}

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// GetAppliedTodayCount implements Datastore.
func (d *sqlite) GetAppliedTodayCount(ctx context.Context) (int, error) {
	row := d.db.QueryRowContext(ctx, `SELECT count FROM applied_counts WHERE date = date('now')`)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return 0, nil
		}

		return 0, row.Err()
	}

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (d *sqlite) IncAppliedTodayCount(ctx context.Context, platform string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	query := "SELECT count FROM applied_counts WHERE platform = ? AND date = date('now')"
	row := tx.QueryRow(query, platform)

	var existingCount int
	err = row.Scan(&existingCount)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	newCount := 1
	if existingCount > 0 {
		newCount += existingCount
	}

	insertQuery := "INSERT OR REPLACE INTO applied_counts (platform, date, count) VALUES (?, date('now'), ?)"
	_, err = tx.Exec(insertQuery, platform, newCount)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetUnappliedJobPosting returns a job posting that has not been applied to yet.
// If there are no unapplied job postings, nil is returned.
func (d *sqlite) GetUnappliedJobPosting(ctx context.Context) (*JobPosting, error) {
	row := d.db.QueryRowContext(ctx, `
		SELECT platform, id, url, job_title, company, applied
		FROM job_postings
		WHERE applied = 0
		LIMIT 1
	`)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, nil
		}
		return nil, row.Err()
	}

	var jobPosting JobPosting
	if err := row.Scan(
		&jobPosting.Platform,
		&jobPosting.ID,
		&jobPosting.Url,
		&jobPosting.Title,
		&jobPosting.Company,
		&jobPosting.Applied,
	); err != nil {
		return nil, err
	}

	return &jobPosting, nil
}

// IncAppliedCountByCompany increases the applied count for a company.
func (d *sqlite) IncAppliedCountByCompany(ctx context.Context, name string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	query := "SELECT count FROM applied_counts_by_company WHERE name = ? AND date = date('now')"
	row := tx.QueryRow(query, name)

	var existingCount int
	err = row.Scan(&existingCount)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	newCount := 1
	if existingCount > 0 {
		newCount += existingCount
	}

	insertQuery := "INSERT OR REPLACE INTO applied_counts_by_company (name, date, count) VALUES (?, date('now'), ?)"
	_, err = tx.Exec(insertQuery, name, newCount)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *sqlite) InsertJobPosting(ctx context.Context, jobPosting *JobPosting) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO job_postings (platform, id, url, job_title, company, applied)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec(
		jobPosting.Platform,
		jobPosting.ID,
		jobPosting.Url,
		jobPosting.Title,
		jobPosting.Company,
		jobPosting.Applied,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *sqlite) Close() error {
	return d.db.Close()
}

// NewDatastore creates a new SQLite-based datastore.
func NewSqliteDatastore(dbFile string) (Datastore, error) {
	if dbFile == "" {
		cfgFile := viper.ConfigFileUsed()
		if cfgFile == "" {
			return nil, errors.New("config file not found")
		}
		dbFile = strings.TrimSuffix(cfgFile, filepath.Ext(cfgFile)) + ".db"
		log.Info().Str("db_file", dbFile).Msg("using default db file")
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	// Create the table if it doesn't exist
	_, err = db.Exec(createSqliteTableQuery)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &sqlite{db: db}, nil
}
