// Code generated by sqlc. DO NOT EDIT.
// source: query.sql

package dcrdb

import (
	"context"
	"database/sql"

	"github.com/jackc/pgconn"
)

const existsAggregatedSignature = `-- name: ExistsAggregatedSignature :one
SELECT EXISTS(SELECT 1 FROM decryptor.aggregated_signature WHERE signed_hash = $1)
`

func (q *Queries) ExistsAggregatedSignature(ctx context.Context, signedHash []byte) (bool, error) {
	row := q.db.QueryRow(ctx, existsAggregatedSignature, signedHash)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const getAggregatedSignature = `-- name: GetAggregatedSignature :one
SELECT epoch_id, signed_hash, signers_bitfield, signature FROM decryptor.aggregated_signature
WHERE signed_hash = $1
`

func (q *Queries) GetAggregatedSignature(ctx context.Context, signedHash []byte) (DecryptorAggregatedSignature, error) {
	row := q.db.QueryRow(ctx, getAggregatedSignature, signedHash)
	var i DecryptorAggregatedSignature
	err := row.Scan(
		&i.EpochID,
		&i.SignedHash,
		&i.SignersBitfield,
		&i.Signature,
	)
	return i, err
}

const getCipherBatch = `-- name: GetCipherBatch :one
SELECT epoch_id, transactions FROM decryptor.cipher_batch
WHERE epoch_id = $1
`

func (q *Queries) GetCipherBatch(ctx context.Context, epochID []byte) (DecryptorCipherBatch, error) {
	row := q.db.QueryRow(ctx, getCipherBatch, epochID)
	var i DecryptorCipherBatch
	err := row.Scan(&i.EpochID, &i.Transactions)
	return i, err
}

const getDecryptionKey = `-- name: GetDecryptionKey :one
SELECT epoch_id, key FROM decryptor.decryption_key
WHERE epoch_id = $1
`

func (q *Queries) GetDecryptionKey(ctx context.Context, epochID []byte) (DecryptorDecryptionKey, error) {
	row := q.db.QueryRow(ctx, getDecryptionKey, epochID)
	var i DecryptorDecryptionKey
	err := row.Scan(&i.EpochID, &i.Key)
	return i, err
}

const getDecryptionSignature = `-- name: GetDecryptionSignature :one
SELECT epoch_id, signed_hash, signers_bitfield, signature FROM decryptor.decryption_signature
WHERE epoch_id = $1 AND signers_bitfield = $2
`

type GetDecryptionSignatureParams struct {
	EpochID         []byte
	SignersBitfield []byte
}

func (q *Queries) GetDecryptionSignature(ctx context.Context, arg GetDecryptionSignatureParams) (DecryptorDecryptionSignature, error) {
	row := q.db.QueryRow(ctx, getDecryptionSignature, arg.EpochID, arg.SignersBitfield)
	var i DecryptorDecryptionSignature
	err := row.Scan(
		&i.EpochID,
		&i.SignedHash,
		&i.SignersBitfield,
		&i.Signature,
	)
	return i, err
}

const getDecryptionSignatures = `-- name: GetDecryptionSignatures :many
SELECT epoch_id, signed_hash, signers_bitfield, signature FROM decryptor.decryption_signature
WHERE epoch_id = $1 AND signed_hash = $2
`

type GetDecryptionSignaturesParams struct {
	EpochID    []byte
	SignedHash []byte
}

func (q *Queries) GetDecryptionSignatures(ctx context.Context, arg GetDecryptionSignaturesParams) ([]DecryptorDecryptionSignature, error) {
	rows, err := q.db.Query(ctx, getDecryptionSignatures, arg.EpochID, arg.SignedHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DecryptorDecryptionSignature
	for rows.Next() {
		var i DecryptorDecryptionSignature
		if err := rows.Scan(
			&i.EpochID,
			&i.SignedHash,
			&i.SignersBitfield,
			&i.Signature,
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

const getDecryptorIndex = `-- name: GetDecryptorIndex :one
SELECT index
FROM decryptor.decryptor_set_member
WHERE activation_block_number <= $1 AND address = $2
ORDER BY activation_block_number DESC LIMIT 1
`

type GetDecryptorIndexParams struct {
	ActivationBlockNumber int64
	Address               string
}

func (q *Queries) GetDecryptorIndex(ctx context.Context, arg GetDecryptorIndexParams) (int32, error) {
	row := q.db.QueryRow(ctx, getDecryptorIndex, arg.ActivationBlockNumber, arg.Address)
	var index int32
	err := row.Scan(&index)
	return index, err
}

const getDecryptorKey = `-- name: GetDecryptorKey :one
SELECT bls_public_key FROM decryptor.decryptor_identity WHERE address = (
    SELECT address FROM decryptor.decryptor_set_member
    WHERE index = $1 AND activation_block_number <= $2
    ORDER BY activation_block_number DESC LIMIT 1
)
`

type GetDecryptorKeyParams struct {
	Index                 int32
	ActivationBlockNumber int64
}

func (q *Queries) GetDecryptorKey(ctx context.Context, arg GetDecryptorKeyParams) ([]byte, error) {
	row := q.db.QueryRow(ctx, getDecryptorKey, arg.Index, arg.ActivationBlockNumber)
	var bls_public_key []byte
	err := row.Scan(&bls_public_key)
	return bls_public_key, err
}

const getDecryptorSet = `-- name: GetDecryptorSet :many
SELECT
    member.activation_block_number,
    member.index,
    member.address,
    identity.bls_public_key
FROM (
    SELECT
        activation_block_number,
        index,
        address
    FROM decryptor.decryptor_set_member
    WHERE activation_block_number = (
        SELECT
            m.activation_block_number
        FROM decryptor.decryptor_set_member AS m
        WHERE m.activation_block_number <= $1
        ORDER BY m.activation_block_number DESC
        LIMIT 1
    )
) AS member
LEFT OUTER JOIN decryptor.decryptor_identity AS identity
ON member.address = identity.address
ORDER BY index
`

type GetDecryptorSetRow struct {
	ActivationBlockNumber int64
	Index                 int32
	Address               string
	BlsPublicKey          []byte
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

const getEonPublicKey = `-- name: GetEonPublicKey :one
SELECT eon_public_key
FROM decryptor.eon_public_key
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1
`

func (q *Queries) GetEonPublicKey(ctx context.Context, activationBlockNumber int64) ([]byte, error) {
	row := q.db.QueryRow(ctx, getEonPublicKey, activationBlockNumber)
	var eon_public_key []byte
	err := row.Scan(&eon_public_key)
	return eon_public_key, err
}

const getEventSyncProgress = `-- name: GetEventSyncProgress :one
SELECT id, next_block_number, next_log_index FROM decryptor.event_sync_progress LIMIT 1
`

func (q *Queries) GetEventSyncProgress(ctx context.Context) (DecryptorEventSyncProgress, error) {
	row := q.db.QueryRow(ctx, getEventSyncProgress)
	var i DecryptorEventSyncProgress
	err := row.Scan(&i.ID, &i.NextBlockNumber, &i.NextLogIndex)
	return i, err
}

const getKeyperSet = `-- name: GetKeyperSet :one
SELECT (
    activation_block_number,
    keypers,
    threshold
) FROM decryptor.keyper_set
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1
`

func (q *Queries) GetKeyperSet(ctx context.Context, activationBlockNumber sql.NullInt64) (interface{}, error) {
	row := q.db.QueryRow(ctx, getKeyperSet, activationBlockNumber)
	var column_1 interface{}
	err := row.Scan(&column_1)
	return column_1, err
}

const getMeta = `-- name: GetMeta :one
SELECT key, value FROM decryptor.meta_inf WHERE key = $1
`

func (q *Queries) GetMeta(ctx context.Context, key string) (DecryptorMetaInf, error) {
	row := q.db.QueryRow(ctx, getMeta, key)
	var i DecryptorMetaInf
	err := row.Scan(&i.Key, &i.Value)
	return i, err
}

const insertAggregatedSignature = `-- name: InsertAggregatedSignature :execresult
INSERT INTO decryptor.aggregated_signature (
    epoch_id, signed_hash, signers_bitfield, signature
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT DO NOTHING
`

type InsertAggregatedSignatureParams struct {
	EpochID         []byte
	SignedHash      []byte
	SignersBitfield []byte
	Signature       []byte
}

func (q *Queries) InsertAggregatedSignature(ctx context.Context, arg InsertAggregatedSignatureParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertAggregatedSignature,
		arg.EpochID,
		arg.SignedHash,
		arg.SignersBitfield,
		arg.Signature,
	)
}

const insertCipherBatch = `-- name: InsertCipherBatch :execresult
INSERT INTO decryptor.cipher_batch (
    epoch_id, transactions
) VALUES (
    $1, $2
)
ON CONFLICT DO NOTHING
`

type InsertCipherBatchParams struct {
	EpochID      []byte
	Transactions [][]byte
}

func (q *Queries) InsertCipherBatch(ctx context.Context, arg InsertCipherBatchParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertCipherBatch, arg.EpochID, arg.Transactions)
}

const insertDecryptionKey = `-- name: InsertDecryptionKey :execresult
INSERT INTO decryptor.decryption_key (
    epoch_id, key
) VALUES (
    $1, $2
)
ON CONFLICT DO NOTHING
`

type InsertDecryptionKeyParams struct {
	EpochID []byte
	Key     []byte
}

func (q *Queries) InsertDecryptionKey(ctx context.Context, arg InsertDecryptionKeyParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertDecryptionKey, arg.EpochID, arg.Key)
}

const insertDecryptionSignature = `-- name: InsertDecryptionSignature :execresult
INSERT INTO decryptor.decryption_signature (
    epoch_id, signed_hash, signers_bitfield, signature
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT DO NOTHING
`

type InsertDecryptionSignatureParams struct {
	EpochID         []byte
	SignedHash      []byte
	SignersBitfield []byte
	Signature       []byte
}

func (q *Queries) InsertDecryptionSignature(ctx context.Context, arg InsertDecryptionSignatureParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertDecryptionSignature,
		arg.EpochID,
		arg.SignedHash,
		arg.SignersBitfield,
		arg.Signature,
	)
}

const insertDecryptorIdentity = `-- name: InsertDecryptorIdentity :exec
INSERT INTO decryptor.decryptor_identity (
    address, bls_public_key
) VALUES (
    $1, $2
)
`

type InsertDecryptorIdentityParams struct {
	Address      string
	BlsPublicKey []byte
}

func (q *Queries) InsertDecryptorIdentity(ctx context.Context, arg InsertDecryptorIdentityParams) error {
	_, err := q.db.Exec(ctx, insertDecryptorIdentity, arg.Address, arg.BlsPublicKey)
	return err
}

const insertDecryptorSetMember = `-- name: InsertDecryptorSetMember :exec
INSERT INTO decryptor.decryptor_set_member (
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

const insertEonPublicKey = `-- name: InsertEonPublicKey :exec
INSERT INTO decryptor.eon_public_key (
    activation_block_number,
    eon_public_key
) VALUES (
    $1, $2
)
`

type InsertEonPublicKeyParams struct {
	ActivationBlockNumber int64
	EonPublicKey          []byte
}

func (q *Queries) InsertEonPublicKey(ctx context.Context, arg InsertEonPublicKeyParams) error {
	_, err := q.db.Exec(ctx, insertEonPublicKey, arg.ActivationBlockNumber, arg.EonPublicKey)
	return err
}

const insertKeyperSet = `-- name: InsertKeyperSet :exec
INSERT INTO decryptor.keyper_set (
    activation_block_number,
    keypers,
    threshold
) VALUES (
    $1, $2, $3
)
`

type InsertKeyperSetParams struct {
	ActivationBlockNumber sql.NullInt64
	Keypers               []string
	Threshold             int32
}

func (q *Queries) InsertKeyperSet(ctx context.Context, arg InsertKeyperSetParams) error {
	_, err := q.db.Exec(ctx, insertKeyperSet, arg.ActivationBlockNumber, arg.Keypers, arg.Threshold)
	return err
}

const insertMeta = `-- name: InsertMeta :exec
INSERT INTO decryptor.meta_inf (key, value) VALUES ($1, $2)
`

type InsertMetaParams struct {
	Key   string
	Value string
}

func (q *Queries) InsertMeta(ctx context.Context, arg InsertMetaParams) error {
	_, err := q.db.Exec(ctx, insertMeta, arg.Key, arg.Value)
	return err
}

const updateEventSyncProgress = `-- name: UpdateEventSyncProgress :exec
INSERT INTO decryptor.event_sync_progress (next_block_number, next_log_index)
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
