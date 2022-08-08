package chain

import (
	"errors"
	"fmt"

	"github.com/ChainSafe/log15"
	stafiHubXLedgerTypes "github.com/stafihub/stafihub/x/ledger/types"
	stafiHubXRBankTypes "github.com/stafihub/stafihub/x/rbank/types"
)

var (
	ErrorTerminated = errors.New("terminated")
)

type Chain struct {
	conn        *Connection
	listener    *Listener // The listener of this chain
	stop        chan<- struct{}
	initialized bool
}

func NewChain() *Chain {
	return &Chain{}
}

func (c *Chain) Initialize(option *ConfigOption, logger log15.Logger, sysErr chan<- error) error {
	stop := make(chan struct{})

	conn, err := NewConnection(option, logger)
	if err != nil {
		return err
	}

	bs, err := NewBlockstore(option.BlockstorePath, conn.BlockStoreUseAddress())
	if err != nil {
		return err
	}

	var startBlk uint64
	startBlk, err = StartBlock(bs, uint64(option.StartBlock))
	if err != nil {
		return err
	}

	l := NewListener(option.DropInfos, startBlk, bs, conn, logger, stop, sysErr)

	c.listener = l
	c.conn = conn
	c.initialized = true
	c.stop = stop
	return nil
}

func (c *Chain) Start() error {
	if !c.initialized {
		return fmt.Errorf("chain must be initialized with Initialize()")
	}
	return c.listener.start()
}

// stop will stop handler and listener
func (c *Chain) Stop() {
	close(c.stop)
}

func (c *Chain) GetRParams(denom string) (*stafiHubXLedgerTypes.QueryGetRParamsResponse, error) {
	return c.conn.client.QueryRParams(denom)
}

func (c *Chain) GetPoolDetail(denom, pool string) (*stafiHubXLedgerTypes.QueryGetPoolDetailResponse, error) {
	return c.conn.client.QueryPoolDetail(denom, pool)
}

func (c *Chain) GetPools(denom string) (*stafiHubXLedgerTypes.QueryBondedPoolsByDenomResponse, error) {
	return c.conn.client.QueryPools(denom)
}

func (c *Chain) GetAddressPrefix(denom string) (*stafiHubXRBankTypes.QueryAddressPrefixResponse, error) {
	return c.conn.client.QueryAddressPrefix(denom)
}
