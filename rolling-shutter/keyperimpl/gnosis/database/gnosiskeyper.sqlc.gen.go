// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: gnosiskeyper.sql

package database

import (
	"context"
	"database/sql"

	"github.com/jackc/pgconn"
)

const getConsensusTxPointer = `-- name: GetConsensusTxPointer :one
SELECT consensus, consensus_block FROM tx_pointer LIMIT 1
`

type GetConsensusTxPointerRow struct {
	Consensus      sql.NullInt64
	ConsensusBlock int64
}

func (q *Queries) GetConsensusTxPointer(ctx context.Context) (GetConsensusTxPointerRow, error) {
	row := q.db.QueryRow(ctx, getConsensusTxPointer)
	var i GetConsensusTxPointerRow
	err := row.Scan(&i.Consensus, &i.ConsensusBlock)
	return i, err
}

const getLocalTxPointer = `-- name: GetLocalTxPointer :one
SELECT local, local_block FROM tx_pointer LIMIT 1
`

type GetLocalTxPointerRow struct {
	Local      sql.NullInt64
	LocalBlock int64
}

func (q *Queries) GetLocalTxPointer(ctx context.Context) (GetLocalTxPointerRow, error) {
	row := q.db.QueryRow(ctx, getLocalTxPointer)
	var i GetLocalTxPointerRow
	err := row.Scan(&i.Local, &i.LocalBlock)
	return i, err
}

const getTransactionSubmittedEventCount = `-- name: GetTransactionSubmittedEventCount :one
SELECT event_count FROM transaction_submitted_event_count
WHERE eon = $1
LIMIT 1
`

func (q *Queries) GetTransactionSubmittedEventCount(ctx context.Context, eon int64) (int64, error) {
	row := q.db.QueryRow(ctx, getTransactionSubmittedEventCount, eon)
	var event_count int64
	err := row.Scan(&event_count)
	return event_count, err
}

const getTransactionSubmittedEventsSyncedUntil = `-- name: GetTransactionSubmittedEventsSyncedUntil :one
SELECT block_number FROM transaction_submitted_events_synced_until LIMIT 1
`

func (q *Queries) GetTransactionSubmittedEventsSyncedUntil(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, getTransactionSubmittedEventsSyncedUntil)
	var block_number int64
	err := row.Scan(&block_number)
	return block_number, err
}

const insertTransactionSubmittedEvent = `-- name: InsertTransactionSubmittedEvent :execresult
INSERT INTO transaction_submitted_event (
    index,
    block_number,
    block_hash,
    tx_index,
    log_index,
    eon,
    identity_prefix,
    sender,
    gas_limit
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT DO NOTHING
`

type InsertTransactionSubmittedEventParams struct {
	Index          int64
	BlockNumber    int64
	BlockHash      []byte
	TxIndex        int64
	LogIndex       int64
	Eon            int64
	IdentityPrefix []byte
	Sender         string
	GasLimit       int64
}

func (q *Queries) InsertTransactionSubmittedEvent(ctx context.Context, arg InsertTransactionSubmittedEventParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertTransactionSubmittedEvent,
		arg.Index,
		arg.BlockNumber,
		arg.BlockHash,
		arg.TxIndex,
		arg.LogIndex,
		arg.Eon,
		arg.IdentityPrefix,
		arg.Sender,
		arg.GasLimit,
	)
}

const setTransactionSubmittedEventCount = `-- name: SetTransactionSubmittedEventCount :exec
INSERT INTO transaction_submitted_event_count (eon, event_count)
VALUES ($1, $2)
ON CONFLICT (eon) DO UPDATE
SET event_count = $2
`

type SetTransactionSubmittedEventCountParams struct {
	Eon        int64
	EventCount int64
}

func (q *Queries) SetTransactionSubmittedEventCount(ctx context.Context, arg SetTransactionSubmittedEventCountParams) error {
	_, err := q.db.Exec(ctx, setTransactionSubmittedEventCount, arg.Eon, arg.EventCount)
	return err
}

const setTransactionSubmittedEventsSyncedUntil = `-- name: SetTransactionSubmittedEventsSyncedUntil :exec
INSERT INTO transaction_submitted_events_synced_until (block_number) VALUES ($1)
ON CONFLICT (enforce_one_row) DO UPDATE
SET block_number = $1
`

func (q *Queries) SetTransactionSubmittedEventsSyncedUntil(ctx context.Context, blockNumber int64) error {
	_, err := q.db.Exec(ctx, setTransactionSubmittedEventsSyncedUntil, blockNumber)
	return err
}

const updateConsensusTxPointer = `-- name: UpdateConsensusTxPointer :exec
INSERT INTO tx_pointer (consensus, consensus_block)
VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET consensus = $1, consensus_block = $2
`

type UpdateConsensusTxPointerParams struct {
	Consensus      sql.NullInt64
	ConsensusBlock int64
}

func (q *Queries) UpdateConsensusTxPointer(ctx context.Context, arg UpdateConsensusTxPointerParams) error {
	_, err := q.db.Exec(ctx, updateConsensusTxPointer, arg.Consensus, arg.ConsensusBlock)
	return err
}

const updateLocalTxPointer = `-- name: UpdateLocalTxPointer :exec
INSERT INTO tx_pointer (local, local_block)
VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET local = $1, local_block = $2
`

type UpdateLocalTxPointerParams struct {
	Local      sql.NullInt64
	LocalBlock int64
}

func (q *Queries) UpdateLocalTxPointer(ctx context.Context, arg UpdateLocalTxPointerParams) error {
	_, err := q.db.Exec(ctx, updateLocalTxPointer, arg.Local, arg.LocalBlock)
	return err
}
