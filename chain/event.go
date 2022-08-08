package chain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	stafiHubXLedgerTypes "github.com/stafihub/stafihub/x/ledger/types"
)

var (
	ErrEventAttributeNumberUnMatch = errors.New("ErrEventAttributeNumberTooFew")
	fisDenom                       = "ufis"
)

func (l *Listener) processBlockEvents(currentBlock int64) error {
	if currentBlock%100 == 0 {
		l.log.Debug("processEvents", "blockNum", currentBlock)
	}

	txs, err := l.conn.client.GetBlockTxsWithParseErrSkip(currentBlock)
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
				balanceRes, err := l.conn.client.QueryBalance(bonder, fisDenom, 0)
				if err != nil {
					return err
				}
				if balanceRes.GetBalance().Amount.GT(types.ZeroInt()) {
					l.log.Info(fmt.Sprintf("user %s stake amount: %s, denom: %s, but already have: %sufis, will skip", bonder.String(), amount.String(), denom, balanceRes.GetBalance().Amount.String()))
				} else {
					retry := 0
					var txHash string
					for {
						if retry >= BlockRetryLimit {
							return fmt.Errorf("BroadcastBatchMsg reach retry limit: %s", err)
						}
						txHash, err = l.conn.client.SingleTransferTo(bonder, types.NewCoins(types.NewCoin(fisDenom, dropInfo.DropAmount)))
						if err != nil {
							if strings.Contains(strings.ToLower(err.Error()), "incorrect account sequence") {
								l.log.Warn("BroadcastBatchMsg err will retry", "err", err)
								time.Sleep(BlockRetryInterval)
								retry++
								continue
							} else {
								return err
							}
						}
						break
					}

					retry = 0
					var txRes *types.TxResponse
					for {
						if retry >= BlockRetryLimit {
							return fmt.Errorf("QueryTxByHash reach retry limit: %s", err)
						}
						txRes, err = l.conn.client.QueryTxByHash(txHash)
						if err != nil || txRes.Empty() || txRes.Height == 0 {
							l.log.Warn("QueryTxByHash tx failed will retry query", "err", err, "txRes", txRes)
							time.Sleep(BlockRetryInterval)
							retry++
							continue
						}
						break
					}

					l.log.Info(fmt.Sprintf("user %s liquidity bond amount: %s, denom: %s, drop amount: %sufis txHash: %s success", bonder.String(), amount.String(), denom, dropInfo.DropAmount, txHash))
				}

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
