// Code generated by sqlc. DO NOT EDIT.
// source: observe-query.sql

package dcrdb

import (
	"context"
)

const getChainCollator = `-- name: GetChainCollator :one
SELECT activation_block_number, collator FROM chain_collator
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1
`

func (q *Queries) GetChainCollator(ctx context.Context, activationBlockNumber int64) (ChainCollator, error) {
	row := q.db.QueryRow(ctx, getChainCollator, activationBlockNumber)
	var i ChainCollator
	err := row.Scan(&i.ActivationBlockNumber, &i.Collator)
	return i, err
}

const getDecryptorIdentity = `-- name: GetDecryptorIdentity :one
SELECT address, bls_public_key, bls_signature, signature_valid FROM decryptor_identity
WHERE address = $1
`

func (q *Queries) GetDecryptorIdentity(ctx context.Context, address string) (DecryptorIdentity, error) {
	row := q.db.QueryRow(ctx, getDecryptorIdentity, address)
	var i DecryptorIdentity
	err := row.Scan(
		&i.Address,
		&i.BlsPublicKey,
		&i.BlsSignature,
		&i.SignatureValid,
	)
	return i, err
}

const getDecryptorSet = `-- name: GetDecryptorSet :many
SELECT
    member.activation_block_number,
    member.index,
    member.address,
    identity.bls_public_key,
    identity.bls_signature,
    coalesce(identity.signature_valid, false)
FROM (
    SELECT
        activation_block_number,
        index,
        address
    FROM decryptor_set_member
    WHERE activation_block_number = (
        SELECT
            m.activation_block_number
        FROM decryptor_set_member AS m
        WHERE m.activation_block_number <= $1
        ORDER BY m.activation_block_number DESC
        LIMIT 1
    )
) AS member
LEFT OUTER JOIN decryptor_identity AS identity
ON member.address = identity.address
ORDER BY index
`

type GetDecryptorSetRow struct {
	ActivationBlockNumber int64
	Index                 int32
	Address               string
	BlsPublicKey          []byte
	BlsSignature          []byte
	SignatureValid        bool
}

func (q *Queries) GetDecryptorSet(ctx context.Context, activationBlockNumber int64) ([]GetDecryptorSetRow, error) {
	rows, err := q.db.Query(ctx, getDecryptorSet, activationBlockNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDecryptorSetRow
	for rows.Next() {
		var i GetDecryptorSetRow
		if err := rows.Scan(
			&i.ActivationBlockNumber,
			&i.Index,
			&i.Address,
			&i.BlsPublicKey,
			&i.BlsSignature,
			&i.SignatureValid,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDecryptorSetMember = `-- name: GetDecryptorSetMember :one
SELECT
    m1.activation_block_number,
    m1.index,
    m1.address,
    identity.bls_public_key,
    identity.bls_signature,
    coalesce(identity.signature_valid, false)
FROM (
    SELECT
        m2.activation_block_number,
        m2.index,
        m2.address
    FROM decryptor_set_member AS m2
    WHERE activation_block_number = (
        SELECT
            m3.activation_block_number
        FROM decryptor_set_member AS m3
        WHERE m3.activation_block_number <= $1
        ORDER BY m3.activation_block_number DESC
        LIMIT 1
    ) AND m2.index = $2
) AS m1
LEFT OUTER JOIN decryptor_identity AS identity
ON m1.address = identity.address
ORDER BY index
`

type GetDecryptorSetMemberParams struct {
	ActivationBlockNumber int64
	Index                 int32
}

type GetDecryptorSetMemberRow struct {
	ActivationBlockNumber int64
	Index                 int32
	Address               string
	BlsPublicKey          []byte
	BlsSignature          []byte
	SignatureValid        bool
}

func (q *Queries) GetDecryptorSetMember(ctx context.Context, arg GetDecryptorSetMemberParams) (GetDecryptorSetMemberRow, error) {
	row := q.db.QueryRow(ctx, getDecryptorSetMember, arg.ActivationBlockNumber, arg.Index)
	var i GetDecryptorSetMemberRow
	err := row.Scan(
		&i.ActivationBlockNumber,
		&i.Index,
		&i.Address,
		&i.BlsPublicKey,
		&i.BlsSignature,
		&i.SignatureValid,
	)
	return i, err
}

const getEventSyncProgress = `-- name: GetEventSyncProgress :one
SELECT next_block_number, next_log_index FROM event_sync_progress LIMIT 1
`

type GetEventSyncProgressRow struct {
	NextBlockNumber int32
	NextLogIndex    int32
}

func (q *Queries) GetEventSyncProgress(ctx context.Context) (GetEventSyncProgressRow, error) {
	row := q.db.QueryRow(ctx, getEventSyncProgress)
	var i GetEventSyncProgressRow
	err := row.Scan(&i.NextBlockNumber, &i.NextLogIndex)
	return i, err
}

const getKeyperSet = `-- name: GetKeyperSet :one
SELECT (
    activation_block_number,
    keypers,
    threshold
) FROM keyper_set
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1
`

func (q *Queries) GetKeyperSet(ctx context.Context, activationBlockNumber int64) (interface{}, error) {
	row := q.db.QueryRow(ctx, getKeyperSet, activationBlockNumber)
	var column_1 interface{}
	err := row.Scan(&column_1)
	return column_1, err
}

const insertChainCollator = `-- name: InsertChainCollator :exec
INSERT INTO chain_collator (activation_block_number, collator)
VALUES ($1, $2)
`

type InsertChainCollatorParams struct {
	ActivationBlockNumber int64
	Collator              string
}

func (q *Queries) InsertChainCollator(ctx context.Context, arg InsertChainCollatorParams) error {
	_, err := q.db.Exec(ctx, insertChainCollator, arg.ActivationBlockNumber, arg.Collator)
	return err
}

const insertDecryptorIdentity = `-- name: InsertDecryptorIdentity :exec
INSERT INTO decryptor_identity (
    address, bls_public_key, bls_signature, signature_valid
) VALUES (
    $1, $2, $3, $4
)
`

type InsertDecryptorIdentityParams struct {
	Address        string
	BlsPublicKey   []byte
	BlsSignature   []byte
	SignatureValid bool
}

func (q *Queries) InsertDecryptorIdentity(ctx context.Context, arg InsertDecryptorIdentityParams) error {
	_, err := q.db.Exec(ctx, insertDecryptorIdentity,
		arg.Address,
		arg.BlsPublicKey,
		arg.BlsSignature,
		arg.SignatureValid,
	)
	return err
}

const insertDecryptorSetMember = `-- name: InsertDecryptorSetMember :exec
INSERT INTO decryptor_set_member (
    activation_block_number, index, address
) VALUES (
    $1, $2, $3
)
`

type InsertDecryptorSetMemberParams struct {
	ActivationBlockNumber int64
	Index                 int32
	Address               string
}

func (q *Queries) InsertDecryptorSetMember(ctx context.Context, arg InsertDecryptorSetMemberParams) error {
	_, err := q.db.Exec(ctx, insertDecryptorSetMember, arg.ActivationBlockNumber, arg.Index, arg.Address)
	return err
}

const insertKeyperSet = `-- name: InsertKeyperSet :exec
INSERT INTO keyper_set (
    event_index,
    activation_block_number,
    keypers,
    threshold
) VALUES (
    $1, $2, $3, $4
)
`

type InsertKeyperSetParams struct {
	EventIndex            int64
	ActivationBlockNumber int64
	Keypers               []string
	Threshold             int32
}

func (q *Queries) InsertKeyperSet(ctx context.Context, arg InsertKeyperSetParams) error {
	_, err := q.db.Exec(ctx, insertKeyperSet,
		arg.EventIndex,
		arg.ActivationBlockNumber,
		arg.Keypers,
		arg.Threshold,
	)
	return err
}

const updateEventSyncProgress = `-- name: UpdateEventSyncProgress :exec
INSERT INTO event_sync_progress (next_block_number, next_log_index)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
    SET next_block_number = $1,
        next_log_index = $2
`

type UpdateEventSyncProgressParams struct {
	NextBlockNumber int32
	NextLogIndex    int32
}

func (q *Queries) UpdateEventSyncProgress(ctx context.Context, arg UpdateEventSyncProgressParams) error {
	_, err := q.db.Exec(ctx, updateEventSyncProgress, arg.NextBlockNumber, arg.NextLogIndex)
	return err
}
