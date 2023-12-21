package optimism

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/database"
	shopclient "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync"
	shopevent "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

var ErrParseKeyperSet = errors.New("can't parse KeyperSet")

type Keyper struct {
	core   *keyper.KeyperCore
	dbpool *pgxpool.Pool
	config *config.Config

	trigger chan<- *broker.Event[*epochkghandler.DecryptionTrigger]
}

func New(c *config.Config) (*Keyper, error) { //nolint:unparam
	return &Keyper{
		config: c,
	}, nil
}

func (kpr *Keyper) Start(ctx context.Context, runner service.Runner) error {
	var err error

	dbpool, err := db.Connect(ctx, runner, kpr.config.DatabaseURL, database.Definition.Name())
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	// TODO: the new latest block handler function will put values into this channel
	trigger := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])
	kpr.trigger = trigger

	// TODO: this will be more generic, since we don't need the contract deployments for
	// the keyper core
	ethConfig := configuration.NewEthnodeConfig()
	ethConfig.EthereumURL = kpr.config.Optimism.JSONRPCURL
	ethConfig.PrivateKey = kpr.config.Optimism.PrivateKey
	kpr.core, err = keyper.New(
		&kprconfig.Config{
			InstanceID:        kpr.config.InstanceID,
			DatabaseURL:       kpr.config.DatabaseURL,
			HTTPEnabled:       kpr.config.HTTPEnabled,
			HTTPListenAddress: kpr.config.HTTPListenAddress,
			P2P:               kpr.config.P2P,
			Ethereum:          ethConfig,
			Shuttermint:       kpr.config.Shuttermint,
			Metrics:           kpr.config.Metrics,
		},
		trigger,
		keyper.WithDBPool(dbpool),
		keyper.NoBroadcastEonPublicKey(),
		keyper.WithEonPublicKeyHandler(kpr.newEonPublicKey),
	)
	if err != nil {
		return errors.Wrap(err, "can't instantiate keyper core")
	}
	// TODO: wrap the logger and pass in
	l2Client, err := shopclient.NewShutterL2Client(
		ctx,
		shopclient.WithClientURL(kpr.config.Optimism.JSONRPCURL),
		shopclient.WithSyncNewBlock(kpr.newBlock),
		shopclient.WithSyncNewKeyperSet(kpr.newKeyperSet),
	)
	// TODO: how to deal with polling past state? (sounds like a big addition to the l2Client)
	if err != nil {
		return err
	}
	return runner.StartService(kpr.core, l2Client)
}

func (kpr *Keyper) newBlock(_ context.Context, ev *shopevent.LatestBlock) error {
	log.Info().
		Int64("number", ev.Number.Int64()).
		Str("hash", ev.BlockHash.Hex()).
		Msg("new latest unsafe head on L2")

	// TODO: sanity checks

	idPreimage := identitypreimage.BigToIdentityPreimage(ev.Number)
	trig := &epochkghandler.DecryptionTrigger{
		BlockNumber:       ev.Number.Uint64(),
		IdentityPreimages: []identitypreimage.IdentityPreimage{idPreimage},
	}

	// TODO: what to do if this blocks?
	kpr.trigger <- broker.NewEvent(trig)
	return nil
}

func (kpr *Keyper) newKeyperSet(ctx context.Context, ev *shopevent.KeyperSet) error {
	log.Info().
		Uint64("activation-block", ev.ActivationBlock).
		Msg("new keyper set added")
	return kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		db := obskeyper.New(tx)

		keyperConfigIndex, err := medley.Uint64ToInt64Safe(ev.Eon)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		activationBlockNumber, err := medley.Uint64ToInt64Safe(ev.ActivationBlock)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		threshold, err := medley.Uint64ToInt64Safe(ev.Threshold)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		// XXX: does this work when the memberset is empty?
		return db.InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
			KeyperConfigIndex:     keyperConfigIndex,
			ActivationBlockNumber: activationBlockNumber,
			Keypers:               shdb.EncodeAddresses(ev.Members),
			Threshold:             int32(threshold),
		})
	})
}

func (kpr *Keyper) newEonPublicKey(ctx context.Context, pk keyper.EonPublicKey) error {
	// TODO: post the public key to the contract
	return nil
}
