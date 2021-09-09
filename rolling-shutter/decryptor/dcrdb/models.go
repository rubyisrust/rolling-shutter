// Code generated by sqlc. DO NOT EDIT.

package dcrdb

import ()

type DecryptorCipherBatch struct {
	EpochID int64
	Data    []byte
}

type DecryptorDecryptionKey struct {
	EpochID int64
	Key     []byte
}

type DecryptorDecryptionSignature struct {
	EpochID     int64
	SignedHash  []byte
	SignerIndex int64
	Signature   []byte
}
