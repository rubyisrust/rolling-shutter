package shtx

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

const (
	CipherTransactionType    uint8 = 0x50
	PlaintextTransactionType uint8 = 0x51
)

var (
	ErrInputTooShort          = errors.New("input too short")
	ErrUnknownTransactionType = errors.New("unknown transaction type")
)

type Transaction interface {
	Type() uint8
	Encode() ([]byte, error)
	EnvelopeSigner() (common.Address, error)
}

func Decode(input []byte) (Transaction, error) {
	if len(input) == 0 {
		return nil, ErrInputTooShort
	}

	typePrefix := input[0]
	payload := input[1:]

	switch typePrefix {
	case CipherTransactionType:
		return decodeCipherTransactionPayload(payload)
	case PlaintextTransactionType:
		return decodePlaintextTransactionPayload(payload)
	default:
		return nil, ErrUnknownTransactionType
	}
}
