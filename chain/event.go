package chain

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	stafiHubXLedgerTypes "github.com/stafihub/stafihub/x/ledger/types"
)

var (
	ErrEventAttributeNumberUnMatch = errors.New("ErrEventAttributeNumberTooFew")
)

func (l *Listener) processBlockEvents(currentBlock int64) error {
	if currentBlock%100 == 0 {
		l.log.Debug("processEvents", "blockNum", currentBlock)
	}

	txs, err := l.conn.client.GetBlockTxs(currentBlock)
	if err != nil {
		return fmt.Errorf("client.GetBlockTxs failed: %s", err)
	}
	for _, tx := range txs {
		for _, log := range tx.Logs {
			for _, event := range log.Events {
				err := l.processStringEvents(event, currentBlock)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (l *Listener) processStringEvents(event types.StringEvent, blockNumber int64) error {
	l.log.Debug("processStringEvents", "event", event)

	switch {
	case event.Type == stafiHubXLedgerTypes.EventTypeBondExecuted:
		if len(event.Attributes) != 6 {
			return ErrEventAttributeNumberUnMatch
		}
		denom := event.Attributes[0].Value
		bonder, err := types.AccAddressFromBech32(event.Attributes[1].Value)
		if err != nil {
			return err
		}
		amount, ok := types.NewIntFromString(event.Attributes[4].Value)
		if !ok {
			return fmt.Errorf("amount format not right, amount: %s", event.Attributes[4].Value)
		}

		if dropInfo, exist := l.dropInfos[denom]; !exist {
			l.log.Info(fmt.Sprintf("user %s liquidity bond, but denom %s not support drop, will skip", bonder.String(), denom))
		} else {
			if amount.GTE(dropInfo.MinBondAmount) {
				txHash, err := l.conn.client.SingleTransferTo(bonder, types.NewCoins(types.NewCoin("ufis", dropInfo.DropAmount)))
				if err != nil {
					return err
				}
				l.log.Info(fmt.Sprintf("user %s liquidity bond amount: %s, denom: %s, drop amount: %sufis txHash: %s success", bonder.String(), amount.String(), denom, dropInfo.DropAmount, txHash))
			} else {
				l.log.Info(fmt.Sprintf("user %s stake amount: %s, denom: %s, but less than minBondAmount %s, will skip", bonder.String(), amount.String(), denom, dropInfo.MinBondAmount))
			}
		}

	default:
		return nil
	}

	l.log.Info("find liquidity bond event", "block number", blockNumber)
	return nil
}
