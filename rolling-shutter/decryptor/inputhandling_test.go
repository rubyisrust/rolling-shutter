package decryptor

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func newTestConfig(t *testing.T) Config {
	t.Helper()

	p2pKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	assert.NilError(t, err)
	signingKey, _, err := shbls.RandomKeyPair(rand.Reader)
	assert.NilError(t, err)
	return Config{
		ListenAddress:  nil,
		PeerMultiaddrs: nil,

		DatabaseURL: "",

		P2PKey:      p2pKey,
		SigningKey:  signingKey,
		SignerIndex: 1,

		InstanceID: 123,
	}
}

func TestInsertDecryptionKeyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)
	tkg := medley.NewTestKeyGenerator(t)

	err := db.InsertEonPublicKey(ctx, dcrdb.InsertEonPublicKeyParams{
		StartEpochID: medley.Uint64EpochIDToBytes(0),
		EonPublicKey: tkg.EonPublicKey(0).Marshal(),
	})
	assert.NilError(t, err)

	// send an epoch secret key and check that it's stored in the db
	m := &decryptionKey{
		epochID: 0,
		key:     tkg.EpochSecretKey(0),
	}
	msgs, err := handleDecryptionKeyInput(ctx, config, db, m)
	assert.NilError(t, err)

	mStored, err := db.GetDecryptionKey(ctx, medley.Uint64EpochIDToBytes(m.epochID))
	assert.NilError(t, err)
	assert.Check(t, medley.BytesEpochIDToUint64(mStored.EpochID) == m.epochID)
	keyBytes, _ := m.key.GobEncode()
	assert.Check(t, bytes.Equal(mStored.Key, keyBytes))

	assert.Check(t, len(msgs) == 0)

	// send a wrong epoch secret key (e.g., one for a wrong epoch) and check that there's an error
	m2 := &decryptionKey{
		epochID: 1,
		key:     tkg.EpochSecretKey(2),
	}
	_, err = handleDecryptionKeyInput(ctx, config, db, m2)
	assert.Check(t, err != nil)
}

func TestInsertCipherBatchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)

	m := &cipherBatch{
		EpochID:      100,
		Transactions: [][]byte{[]byte("tx1"), []byte("tx2")},
	}
	msgs, err := handleCipherBatchInput(ctx, config, db, m)
	assert.NilError(t, err)

	mStored, err := db.GetCipherBatch(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	assert.NilError(t, err)
	assert.Check(t, medley.BytesEpochIDToUint64(mStored.EpochID) == m.EpochID)
	assert.DeepEqual(t, mStored.Transactions, m.Transactions)
	assert.Check(t, len(msgs) == 0)

	m2 := &cipherBatch{
		EpochID:      100,
		Transactions: [][]byte{[]byte("tx3")},
	}
	msgs, err = handleCipherBatchInput(ctx, config, db, m2)
	assert.NilError(t, err)

	m2Stored, err := db.GetCipherBatch(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	assert.NilError(t, err)
	assert.DeepEqual(t, m2Stored.Transactions, m.Transactions)

	assert.Check(t, len(msgs) == 0)
}

func TestHandleSignatureIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	config := newTestConfig(t)
	configTwoRequiredSignatures := config
	configTwoRequiredSignatures.RequiredSignatures = 2

	signingKey2, _, err := shbls.RandomKeyPair(rand.Reader)
	assert.NilError(t, err)

	err = db.InsertDecryptorSetMember(ctx, dcrdb.InsertDecryptorSetMemberParams{
		StartEpochID: []byte{0},
		Index:        1,
		Address:      "0xdeadbeef",
	})
	assert.NilError(t, err)
	err = db.InsertDecryptorSetMember(ctx, dcrdb.InsertDecryptorSetMemberParams{
		StartEpochID: []byte{0},
		Index:        2,
		Address:      "0xabcdefabcdef",
	})
	assert.NilError(t, err)

	err = db.InsertDecryptorIdentity(ctx, dcrdb.InsertDecryptorIdentityParams{
		Address:      "0xdeadbeef",
		BlsPublicKey: shbls.SecretToPublicKey(config.SigningKey).Marshal(),
	})
	assert.NilError(t, err)
	err = db.InsertDecryptorIdentity(ctx, dcrdb.InsertDecryptorIdentityParams{
		Address:      "0xabcdefabcdef",
		BlsPublicKey: shbls.SecretToPublicKey(signingKey2).Marshal(),
	})
	assert.NilError(t, err)

	bitfield := makeBitfieldFromIndex(1)
	bitfield2 := makeBitfieldFromIndex(2)
	hash := common.BytesToHash([]byte("Hello"))
	signature := &decryptionSignature{
		epochID:        0,
		instanceID:     config.InstanceID,
		signedHash:     hash,
		signature:      shbls.Sign(hash.Bytes(), config.SigningKey),
		SignerBitfield: bitfield,
	}
	signature2 := &decryptionSignature{
		epochID:        0,
		instanceID:     config.InstanceID,
		signedHash:     hash,
		signature:      shbls.Sign(hash.Bytes(), signingKey2),
		SignerBitfield: bitfield2,
	}

	// The tests are not fully independent since database is not wiped in between each one.
	tests := []struct {
		config  Config
		inputs  []*decryptionSignature
		outputs []*shmsg.AggregatedDecryptionSignature
	}{
		{
			config:  config,
			inputs:  []*decryptionSignature{signature},
			outputs: []*shmsg.AggregatedDecryptionSignature{{InstanceID: config.InstanceID, SignedHash: hash.Bytes(), SignerBitfield: bitfield}},
		},
		{
			config:  configTwoRequiredSignatures,
			inputs:  []*decryptionSignature{signature},
			outputs: []*shmsg.AggregatedDecryptionSignature{nil},
		},
		{
			config: configTwoRequiredSignatures,
			inputs: []*decryptionSignature{signature, signature2},
			outputs: []*shmsg.AggregatedDecryptionSignature{nil, {
				InstanceID: configTwoRequiredSignatures.InstanceID, SignedHash: hash.Bytes(),
				SignerBitfield: makeBitfieldFromArray([]int32{1, 2}),
			}},
		},
	}

	for _, test := range tests {
		for i, input := range test.inputs {
			msgs, err := handleSignatureInput(ctx, test.config, db, input)
			assert.NilError(t, err)
			isOutputNill := test.outputs[i] == nil
			if isOutputNill {
				assert.Check(t, len(msgs) == 0)
			} else {
				assert.Check(t, len(msgs) == 1)
				msg, ok := msgs[0].(*shmsg.AggregatedDecryptionSignature)
				assert.Check(t, ok, "wrong message type")
				assert.Equal(t, msg.InstanceID, test.outputs[i].InstanceID)
				assert.Check(t, bytes.Equal(msg.SignedHash, test.outputs[i].SignedHash))
				assert.Check(t, bytes.Equal(msg.SignerBitfield, test.outputs[i].SignerBitfield))
			}
		}
	}
}

func TestHandleEpochIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)
	config.RequiredSignatures = 2 // prevent generation of polluting signatures
	tkg := medley.NewTestKeyGenerator(t)

	err := db.InsertEonPublicKey(ctx, dcrdb.InsertEonPublicKeyParams{
		StartEpochID: medley.Uint64EpochIDToBytes(0),
		EonPublicKey: tkg.EonPublicKey(0).Marshal(),
	})
	assert.NilError(t, err)

	cipherBatchMsg := &cipherBatch{
		EpochID:      0,
		Transactions: [][]byte{[]byte("tx1")},
	}
	msgs, err := handleCipherBatchInput(ctx, config, db, cipherBatchMsg)
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)

	keyMsg := &decryptionKey{
		epochID: 0,
		key:     tkg.EpochSecretKey(0),
	}
	msgs, err = handleDecryptionKeyInput(ctx, config, db, keyMsg)
	assert.NilError(t, err)

	storedDecryptionKey,
		err := db.GetDecryptionSignature(ctx, dcrdb.GetDecryptionSignatureParams{
		EpochID:         medley.Uint64EpochIDToBytes(cipherBatchMsg.EpochID),
		SignersBitfield: makeBitfieldFromIndex(config.SignerIndex),
	})
	assert.NilError(t, err)

	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*shmsg.AggregatedDecryptionSignature)
	assert.Check(t, ok, "wrong message type")
	assert.Equal(
		t,
		medley.BytesEpochIDToUint64(storedDecryptionKey.EpochID),
		msg.EpochID,
	)
	assert.Check(t, bytes.Equal(storedDecryptionKey.SignedHash, msg.SignedHash))
	assert.Check(t, bytes.Equal(storedDecryptionKey.Signature, msg.AggregatedSignature))
}
