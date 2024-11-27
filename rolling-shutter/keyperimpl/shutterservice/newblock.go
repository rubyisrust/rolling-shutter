package shutterservice

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	servicedatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
)

func (kpr *Keyper) processNewBlock(ctx context.Context, ev *syncevent.LatestBlock) error {
	if kpr.registrySyncer != nil {
		if err := kpr.registrySyncer.Sync(ctx, ev.Header); err != nil {
			return err
		}
	}
	return kpr.maybeTriggerDecryption(ctx, ev)
}

// maybeTriggerDecryption triggers decryption for the identities registered if
// - it hasn't been triggered for thos identities before and
// - the keyper is part of the corresponding keyper set.
func (kpr *Keyper) maybeTriggerDecryption(ctx context.Context, block *syncevent.LatestBlock) error {
	if kpr.latestTriggeredTime != nil && block.Header.Time <= *kpr.latestTriggeredTime {
		return nil
	}

	lastTriggeredTime := kpr.latestTriggeredTime
	kpr.latestTriggeredTime = &block.Header.Time

	fmt.Println("--------------")
	fmt.Println(block.Header.Time)
	fmt.Println("--------------")

	serviceDB := servicedatabase.New(kpr.dbpool)
	nonTriggeredEvents, err := serviceDB.GetNotDecryptedIdentityRegisteredEvents(ctx, int64(*lastTriggeredTime))
	if err != nil && err != pgx.ErrNoRows {
		// pgx.ErrNoRows is expected if we're not part of the keyper set (which is checked later).
		// That's because non-keypers don't sync transaction submitted events.
		return errors.Wrap(err, "failed to query non decrypted identity registered events from db")
	}

	obsDB := obskeyper.New(kpr.dbpool)
	eventsToDecrypt := make([]servicedatabase.IdentityRegisteredEvent, 0)
	for _, event := range nonTriggeredEvents {
		if kpr.shouldTriggerDecryption(ctx, obsDB, event) {
			eventsToDecrypt = append(eventsToDecrypt, event)
		}
	}

	//TODO: decryption needs to be implemented
	return nil
}

func (kpr *Keyper) shouldTriggerDecryption(ctx context.Context, obsDB *obskeyper.Queries, event servicedatabase.IdentityRegisteredEvent) bool {
	nextBlock := event.BlockNumber + 1
	keyperSet, err := obsDB.GetKeyperSet(ctx, nextBlock)
	if err == pgx.ErrNoRows {
		log.Info().
			Int64("block-number", nextBlock).
			Msg("skipping event as no keyper set has been found for it")
		return false
	}
	if err != nil {
		log.Warn().Msgf("%w | failed to query keyper set for block %d", err, nextBlock)
		return false
	}
	// don't trigger if we're not part of the keyper set
	if !keyperSet.Contains(kpr.config.GetAddress()) {
		log.Info().
			Int64("block-number", nextBlock).
			Int64("keyper-set-index", keyperSet.KeyperConfigIndex).
			Str("address", kpr.config.GetAddress().Hex()).
			Msg("skipping slot as not part of keyper set")
		return false
	}
	return true
}
