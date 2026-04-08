package ethereum

import (
	"context"
	"encoding/hex"
	"regexp"
	"strings"
	"time"

	"github.com/dapplink-labs/dapplink-wallet-api/chain"
	"github.com/dapplink-labs/dapplink-wallet-api/chain/evmbase"
	"github.com/dapplink-labs/dapplink-wallet-api/config"
	wallet_api "github.com/dapplink-labs/dapplink-wallet-api/protobuf/wallet-api"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

const (
	ChainName string = "Ethereum"
)

type ChainAdaptor struct {
	ethClient     evmbase.EthClient
	ethDataClient *evmbase.EthData
}

func NewChainAdaptor(conf *config.Config) (chain.IChainAdaptor, error) {
	ethClient, err := evmbase.DialEthClient(context.Background(), conf.WalletNode.Eth.RpcUrl)
	if err != nil {
		log.Error("Dial eth client fail", "err", err)
		return nil, err
	}
	ethDataClient, err := evmbase.NewEthDataClient(conf.WalletNode.Eth.DataApiUrl, conf.WalletNode.Eth.DataApiKey, time.Second*15)
	if err != nil {
		log.Error("new eth data client fail", "err", err)
		return nil, err
	}
	return &ChainAdaptor{
		ethClient:     ethClient,
		ethDataClient: ethDataClient,
	}, nil
}

func (c ChainAdaptor) ConvertAddresses(ctx context.Context, req *wallet_api.ConvertAddressesRequest) (*wallet_api.ConvertAddressesResponse, error) {
	var retAddressList []*wallet_api.Addresses
	for _, publicKeyItem := range req.PublicKey {
		var addressItem *wallet_api.Addresses
		publicKeyBytes, err := hex.DecodeString(publicKeyItem.GetPublicKey())
		if err != nil {
			addressItem = &wallet_api.Addresses{
				Address: "",
			}
			log.Error("decode public key fail", "err", err)
		} else {
			addressItem = &wallet_api.Addresses{
				Address: common.BytesToAddress(crypto.Keccak256(publicKeyBytes[1:])[12:]).String(),
			}
		}
		retAddressList = append(retAddressList, addressItem)
	}
	return &wallet_api.ConvertAddressesResponse{
		Code:    wallet_api.ReturnCode_SUCCESS,
		Msg:     "create batch wallet address success",
		Address: retAddressList,
	}, nil
}

func (c ChainAdaptor) ValidAddresses(ctx context.Context, req *wallet_api.ValidAddressesRequest) (*wallet_api.ValidAddressesResponse, error) {
	var retAddressesValid []*wallet_api.AddressesValid
	for _, addressItem := range req.Addresses {
		var addressesValidItem wallet_api.AddressesValid
		addressesValidItem.Address = addressItem.GetAddress()
		ok := regexp.MustCompile("^[0-9a-fA-F]{40}$").MatchString(addressItem.GetAddress()[2:])
		if len(addressItem.GetAddress()) != 42 || !strings.HasPrefix(addressItem.GetAddress(), "0x") || !ok {
			addressesValidItem.Valid = false
		} else {
			addressesValidItem.Valid = true
		}
		retAddressesValid = append(retAddressesValid, &addressesValidItem)
	}
	return &wallet_api.ValidAddressesResponse{
		Code:         wallet_api.ReturnCode_SUCCESS,
		Msg:          "valid batch wallet address success",
		AddressValid: retAddressesValid,
	}, nil
}

func (c ChainAdaptor) GetLastestBlock(ctx context.Context, req *wallet_api.LastestBlockRequest) (*wallet_api.LastestBlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) GetBlock(ctx context.Context, req *wallet_api.BlockRequest) (*wallet_api.BlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) GetTransactionByHash(ctx context.Context, req *wallet_api.TransactionByHashRequest) (*wallet_api.TransactionByHashResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) GetTransactionByAddress(ctx context.Context, req *wallet_api.TransactionByAddressRequest) (*wallet_api.TransactionByAddressResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) GetAccountBalance(ctx context.Context, req *wallet_api.AccountBalanceRequest) (*wallet_api.AccountBalanceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) SendTransaction(ctx context.Context, req *wallet_api.SendTransactionsRequest) (*wallet_api.SendTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) BuildTransactionSchema(ctx context.Context, request *wallet_api.TransactionSchemaRequest) (*wallet_api.TransactionSchemaResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) BuildUnSignTransaction(ctx context.Context, request *wallet_api.UnSignTransactionRequest) (*wallet_api.UnSignTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) BuildSignedTransaction(ctx context.Context, request *wallet_api.SignedTransactionRequest) (*wallet_api.SignedTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c ChainAdaptor) GetAddressApproveList(ctx context.Context, request *wallet_api.AddressApproveListRequest) (*wallet_api.AddressApproveListResponse, error) {
	//TODO implement me
	panic("implement me")
}
