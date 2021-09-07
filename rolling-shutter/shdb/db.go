package shdb

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

// tableNamesQuery returns the names of all user created tables in the database.
const tableNamesQuery = `
	SELECT table_name
	FROM information_schema.tables
	WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
	AND table_schema NOT LIKE 'pg_toast%'
`

// createDecryptorTables creates the tables for the decryptor db.
const createDecryptorTables = `
	CREATE TABLE IF NOT EXISTS cipher_batch (
		round_id bigint PRIMARY KEY,
		data bytea
	);
	CREATE TABLE IF NOT EXISTS decryption_key (
		round_id bigint PRIMARY KEY,
		key bytea
	);
	CREATE TABLE IF NOT EXISTS decryption_signature (
		round_id bigint,
		signed_hash bytea,
		signer_index bigint,
		signature bytea,
		PRIMARY KEY (round_id, signer_index)
	);
`

// InitDecryptorDB initializes the database of the decryptor. It is assumed that the db is empty.
func InitDecryptorDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	_, err := dbpool.Exec(ctx, createDecryptorTables)
	if err != nil {
		return errors.Wrap(err, "failed to create decryptor tables")
	}
	return nil
}

// ValidateKeyperDB checks that all expected tables exist in the database. If not, it returns an
// error.
func ValidateKeyperDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return validateDB(ctx, dbpool, []string{
		"decryption_trigger",
		"decryption_key_share",
		"decryption_key",
	})
}

// ValidateDecryptorDB checks that all expected tables exist in the database. If not, it returns an
// error.
func ValidateDecryptorDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return validateDB(ctx, dbpool, []string{
		"cipher_batch",
		"decryption_key",
		"decryption_signature",
	})
}

func validateDB(ctx context.Context, dbpool *pgxpool.Pool, requiredTables []string) error {
	requiredTableMap := make(map[string]bool)
	for _, table := range requiredTables {
		requiredTableMap[table] = true
	}

	rows, err := dbpool.Query(ctx, tableNamesQuery)
	if err != nil {
		return errors.Wrap(err, "failed to query table names from db")
	}
	defer rows.Close()

	var tableName string
	for rows.Next() {
		err := rows.Scan(&tableName)
		if err != nil {
			return errors.Wrap(err, "failed to query table names from db")
		}
		delete(requiredTableMap, tableName)
	}
	if rows.Err() != nil {
		return errors.Wrap(rows.Err(), "read table names")
	}

	if len(requiredTableMap) != 0 {
		return errors.New("database misses one or more required table")
	}
	return nil
}
