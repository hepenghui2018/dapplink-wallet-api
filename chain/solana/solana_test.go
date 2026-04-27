package solana

import (
	"context"
	"testing"

	"github.com/dapplink-labs/dapplink-wallet-api/chain"
	"github.com/dapplink-labs/dapplink-wallet-api/config"
	wallet_api "github.com/dapplink-labs/dapplink-wallet-api/protobuf/wallet-api"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
)

const ChainName = "Solana"

func setup() (chain.IChainAdaptor, error) {
	conf, err := config.NewConfig("../../config.yml")
	if err != nil {
		log.Error("load config failed, error:", err)
		return nil, err
	}
	adaptor, err := NewChainAdaptor(conf)
	if err != nil {
		log.Error("create chain adaptor failed, error:", err)
		return nil, err
	}
	return adaptor, nil
}

func Test_ValidAddress(t *testing.T) {
	adaptor, err := setup()
	if err != nil {
		t.Skipf("setup failed: %v", err)
	}

	ctx := context.Background()
	req := &wallet_api.ValidAddressesRequest{
		ConsumerToken: "test-token",
		ChainId:       "solana",
		Network:       "mainnet",
		Addresses: []*wallet_api.Addresses{
			{Address: "9VhPRjzizPY95TyBrve7heeJTZnofgkQYJpLxRSZGZ3H"},
		},
	}

	resp, err := adaptor.ValidAddresses(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, wallet_api.ApiReturnCode_APISUCCESS, resp.Code)
	assert.NotEmpty(t, resp.AddressValid)
	if len(resp.AddressValid) > 0 {
		assert.True(t, resp.AddressValid[0].Valid)
	}
}

/*
 * Old tests have been removed as they use deprecated wallet-chain-account interfaces.
 * New tests should be written using the wallet-api protobuf interfaces.
 *
 * TODO: Write comprehensive tests for:
 * - GetSupportChains
 * - ConvertAddress
 * - ValidAddresses
 * - GetBlockByNumber
 * - GetBlockByHash
 * - GetBlockHeaderByNumber
 * - GetAccount
 * - GetFee
 * - SendTx
 * - GetTxByAddress
 * - GetTxByHash
 * - GetAccountBalance
 * - CreateUnSignTransaction
 * - BuildSignedTransaction
 * - DecodeTransaction
 * - VerifySignedTransaction
 * - GetExtraData
 */
