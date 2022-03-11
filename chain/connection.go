package chain

import (
	"fmt"
	"os"

	"github.com/ChainSafe/log15"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	hubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
)



type Connection struct {
	client *hubClient.Client
	log    log15.Logger
}

func NewConnection(cfgOption *ConfigOption, log log15.Logger) (*Connection, error) {
	fmt.Printf("Will open stafihub wallet from <%s>. \nPlease ", cfgOption.KeystorePath)
	key, err := keyring.New(types.KeyringServiceName(), keyring.BackendFile, cfgOption.KeystorePath, os.Stdin)
	if err != nil {
		return nil, err
	}
	client, err := hubClient.NewClient(key, cfgOption.Account, cfgOption.GasPrice, cfgOption.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("hubClient.NewClient err: %s", err)
	}

	c := Connection{
		client: client,
		log:    log,
	}
	return &c, nil
}

func (c *Connection) BlockStoreUseAddress() string {
	return c.client.GetFromAddress().String()
}
