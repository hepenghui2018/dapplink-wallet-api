package bitcoin

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/dapplink-labs/dapplink-wallet-api/chain/bitcoin/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/dapplink-labs/dapplink-wallet-api/chain"
	base "github.com/dapplink-labs/dapplink-wallet-api/chain/bitcoinbase"
	"github.com/dapplink-labs/dapplink-wallet-api/config"
	wallet_api "github.com/dapplink-labs/dapplink-wallet-api/protobuf/wallet-api"
)

const ChainID = "DappLinkBitcoin"

type ChainAdaptor struct {
	btcClient       *base.BaseClient
	btcDataClient   *base.BaseDataClient
	thirdPartClient *BcClient
}

func NewChainAdaptor(conf *config.Config) (chain.IChainAdaptor, error) {
	baseClient, err := base.NewBaseClient(conf.WalletNode.Btc.RpcUrl, conf.WalletNode.Btc.RpcUser, conf.WalletNode.Btc.RpcPass)
	if err != nil {
		log.Error("new bitcoin rpc client fail", "err", err)
		return nil, err
	}
	baseDataClient, err := base.NewBaseDataClient(conf.WalletNode.Btc.DataApiUrl, conf.WalletNode.Btc.DataApiKey, "BTC", "Bitcoin")
	if err != nil {
		log.Error("new bitcoin data client fail", "err", err)
		return nil, err
	}
	bcClient, err := NewBlockChainClient(conf.WalletNode.Btc.TpApiUrl)
	if err != nil {
		log.Error("new blockchain client fail", "err", err)
		return nil, err
	}
	return &ChainAdaptor{
		btcClient:       baseClient,
		btcDataClient:   baseDataClient,
		thirdPartClient: bcClient,
	}, nil
}

func (c ChainAdaptor) ConvertAddresses(ctx context.Context, req *wallet_api.ConvertAddressesRequest) (*wallet_api.ConvertAddressesResponse, error) {
	var addressList []*wallet_api.Addresses
	for _, publicKeyItem := range req.GetPublicKey() {
		var walletAddress wallet_api.Addresses
		compressedPubKeyBytes, _ := hex.DecodeString(publicKeyItem.PublicKey)
		pubKeyHash := btcutil.Hash160(compressedPubKeyBytes)
		switch req.GetAddressFormat() {
		case "p2pkh":
			p2pkhAddr, err := btcutil.NewAddressPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
			if err != nil {
				log.Error("create p2pkh address fail", "err", err, "pubKeyHash", hex.EncodeToString(pubKeyHash))
				walletAddress.Address = ""
			} else {
				walletAddress.Address = p2pkhAddr.EncodeAddress()
			}
			break
		case "p2wpkh":
			witnessAddr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
			if err != nil {
				log.Error("create p2wpkh fail", "err", err, "pubKeyHash", pubKeyHash)
				walletAddress.Address = ""
			} else {
				walletAddress.Address = witnessAddr.EncodeAddress()
			}
			break
		case "p2sh":
			witnessAddr, _ := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, &chaincfg.MainNetParams)
			script, err := txscript.PayToAddrScript(witnessAddr)
			if err != nil {
				log.Error("pay to address script fail", "err", err, "publickey", publicKeyItem.PublicKey)
				walletAddress.Address = ""
			}
			p2shAddr, err := btcutil.NewAddressScriptHash(script, &chaincfg.MainNetParams)
			if err != nil {
				log.Error("create p2sh address fail", "err", err, "publickey", publicKeyItem.PublicKey)
				walletAddress.Address = ""
			} else {
				walletAddress.Address = p2shAddr.EncodeAddress()
			}
			break
		case "p2tr":
			pubKey, err := btcec.ParsePubKey(compressedPubKeyBytes)
			if err != nil {
				log.Error("parse p2tr public fail", "err", err)
				walletAddress.Address = ""
			}
			taprootPubKey := schnorr.SerializePubKey(pubKey)
			taprootAddr, err := btcutil.NewAddressTaproot(taprootPubKey, &chaincfg.MainNetParams)
			if err != nil {
				log.Error("create p2tr address fail", "err", err, "pubkey", pubKey)
				walletAddress.Address = ""
			} else {
				walletAddress.Address = taprootAddr.EncodeAddress()
			}
			break
		default:
			log.Error("unsupported address format", "format", req.GetAddressFormat())
			walletAddress.Address = ""
		}
		addressList = append(addressList, &walletAddress)
	}
	return &wallet_api.ConvertAddressesResponse{
		Code:    wallet_api.ApiReturnCode_APISUCCESS,
		Msg:     "create address success",
		Address: addressList,
	}, nil
}

func (c ChainAdaptor) ValidAddresses(ctx context.Context, req *wallet_api.ValidAddressesRequest) (*wallet_api.ValidAddressesResponse, error) {
	var addressesValidList []*wallet_api.AddressesValid
	for _, addressItem := range req.GetAddresses() {
		var addrValid wallet_api.AddressesValid
		address, err := btcutil.DecodeAddress(addressItem.Address, &chaincfg.MainNetParams)
		addrValid.Address = addressItem.GetAddress()
		if err != nil || !address.IsForNet(&chaincfg.MainNetParams) {
			addrValid.Valid = false
		} else {
			addrValid.Valid = true
		}
	}
	return &wallet_api.ValidAddressesResponse{
		Code:         wallet_api.ApiReturnCode_APISUCCESS,
		Msg:          "verify address success",
		AddressValid: addressesValidList,
	}, nil
}

func (c ChainAdaptor) GetLastestBlock(ctx context.Context, req *wallet_api.LastestBlockRequest) (*wallet_api.LastestBlockResponse, error) {
	blockInfo, err := c.btcClient.GetBlockChainInfo()
	if err != nil {
		log.Error("Get blockchain info fail", "err", err)
		return nil, err
	}
	return &wallet_api.LastestBlockResponse{
		Code:   wallet_api.ApiReturnCode_APISUCCESS,
		Msg:    "Get lastest block success",
		Height: uint64(blockInfo.Headers),
	}, nil
}

func (c ChainAdaptor) GetBlock(ctx context.Context, req *wallet_api.BlockRequest) (*wallet_api.BlockResponse, error) {
	var params []json.RawMessage
	numBlocksJSON, _ := json.Marshal(req.HashHeight)
	params = []json.RawMessage{numBlocksJSON}
	block, _ := c.btcClient.Client.RawRequest("getblock", params)
	var resultBlock types.BlockData
	err := json.Unmarshal(block, &resultBlock)
	if err != nil {
		log.Error("Unmarshal json fail", "err", err)
	}
	var transactionList []*wallet_api.TransactionList
	for _, txid := range resultBlock.Tx {
		txIdJson, _ := json.Marshal(txid)
		boolJSON, _ := json.Marshal(true)
		dataJSON := []json.RawMessage{txIdJson, boolJSON}
		tx, err := c.btcClient.Client.RawRequest("getrawtransaction", dataJSON)
		if err != nil {
			fmt.Println("get raw transaction fail", "err", err)
		}
		var rawTx types.RawTransactionData
		err = json.Unmarshal(tx, &rawTx)
		if err != nil {
			log.Error("json unmarshal fail", "err", err)
			return nil, err
		}

		var fromList []*wallet_api.FromAddress
		for _, vin := range rawTx.Vin {
			fromItem := &wallet_api.FromAddress{
				Amount:  strconv.Itoa(10),
				Address: vin.ScriptSig.Asm,
			}
			fromList = append(fromList, fromItem)
		}
		var toList []*wallet_api.ToAddress
		for _, vout := range rawTx.Vout {
			toItem := &wallet_api.ToAddress{
				Address: vout.ScriptPubKey.Address,
				Amount:  strconv.FormatInt(int64(vout.Value), 10),
			}
			toList = append(toList, toItem)
		}
		txItem := &wallet_api.TransactionList{
			TxHash: rawTx.Hash,
			Fee:    strconv.Itoa(0),
			Status: 0,
			From:   fromList,
			To:     toList,
		}
		transactionList = append(transactionList, txItem)
	}
	return &wallet_api.BlockResponse{
		Code:         wallet_api.ApiReturnCode_APISUCCESS,
		Msg:          "get block by number success",
		Height:       strconv.FormatUint(resultBlock.Height, 10),
		Hash:         req.HashHeight,
		Transactions: transactionList,
	}, nil
}

func (c ChainAdaptor) GetTransactionByHash(ctx context.Context, req *wallet_api.TransactionByHashRequest) (*wallet_api.TransactionByHashResponse, error) {
	txInfo, err := c.thirdPartClient.GetTransactionsByHash(req.Hash)
	if err != nil {
		return &wallet_api.TransactionByHashResponse{
			Code:        wallet_api.ApiReturnCode_APIERROR,
			Msg:         "get transaction list fail",
			Transaction: nil,
		}, err
	}
	var fromAddrs []*wallet_api.FromAddress
	var toAddrs []*wallet_api.ToAddress
	var direction int32
	for _, inputs := range txInfo.Inputs {
		fromAddrs = append(fromAddrs, &wallet_api.FromAddress{Address: inputs.PrevOut.Addr, Amount: inputs.PrevOut.Value.String()})
	}
	txFee := txInfo.Fee
	for _, out := range txInfo.Out {
		toAddrs = append(toAddrs, &wallet_api.ToAddress{Address: out.Addr, Amount: out.Value.String()})
	}
	tx := &wallet_api.TransactionList{
		TxHash: txInfo.Hash,
		From:   fromAddrs,
		To:     toAddrs,
		Fee:    txFee.String(),
		Status: uint32(wallet_api.TxStatus_Success),
		TxType: uint32(direction),
	}
	return &wallet_api.TransactionByHashResponse{
		Code:        wallet_api.ApiReturnCode_APISUCCESS,
		Msg:         "get transaction by hash success",
		Transaction: tx,
	}, nil
}

func (c ChainAdaptor) GetTransactionByAddress(ctx context.Context, req *wallet_api.TransactionByAddressRequest) (*wallet_api.TransactionByAddressResponse, error) {
	transaction, err := c.thirdPartClient.GetTransactionsByAddress(req.Address, strconv.Itoa(int(req.Page)), strconv.Itoa(int(req.PageSize)))
	if err != nil {
		return &wallet_api.TransactionByAddressResponse{
			Code:        wallet_api.ApiReturnCode_APIERROR,
			Msg:         "get transaction list fail",
			Transaction: nil,
		}, err
	}
	var txList []*wallet_api.TransactionList
	for _, ttxs := range transaction.Txs {
		var fromAddrs []*wallet_api.FromAddress
		var toAddrs []*wallet_api.ToAddress
		var direction int32
		for _, inputs := range ttxs.Inputs {
			fromAddrs = append(fromAddrs, &wallet_api.FromAddress{Address: inputs.PrevOut.Addr, Amount: inputs.PrevOut.Value.String()})
		}
		txFee := ttxs.Fee
		for _, out := range ttxs.Out {
			toAddrs = append(toAddrs, &wallet_api.ToAddress{Address: out.Addr, Amount: out.Value.String()})
		}
		if strings.EqualFold(req.Address, fromAddrs[0].Address) {
			direction = 0
		} else {
			direction = 1
		}
		tx := &wallet_api.TransactionList{
			TxHash: ttxs.Hash,
			From:   fromAddrs,
			To:     toAddrs,
			Fee:    txFee.String(),
			Status: uint32(wallet_api.TxStatus_Success),
			TxType: uint32(direction),
		}
		txList = append(txList, tx)
	}
	return &wallet_api.TransactionByAddressResponse{
		Code:        wallet_api.ApiReturnCode_APISUCCESS,
		Msg:         "get transaction list success",
		Transaction: txList,
	}, nil
}

func (c ChainAdaptor) GetAccountBalance(ctx context.Context, req *wallet_api.AccountBalanceRequest) (*wallet_api.AccountBalanceResponse, error) {
	balance, err := c.thirdPartClient.GetAccountBalance(req.Address)
	if err != nil {
		return &wallet_api.AccountBalanceResponse{
			Code:    wallet_api.ApiReturnCode_APIERROR,
			Msg:     "get btc balance fail",
			Balance: "0",
		}, err
	}
	return &wallet_api.AccountBalanceResponse{
		Code:    wallet_api.ApiReturnCode_APISUCCESS,
		Msg:     "get btc balance success",
		Balance: balance,
	}, nil
}

func (c ChainAdaptor) SendTransaction(ctx context.Context, req *wallet_api.SendTransactionsRequest) (*wallet_api.SendTransactionResponse, error) {
	var txRetList []*wallet_api.RawTransactionReturn
	for _, txItem := range req.RawTx {
		var txRetItem wallet_api.RawTransactionReturn
		r := bytes.NewReader([]byte(txItem.RawTx))
		var msgTx wire.MsgTx
		err := msgTx.Deserialize(r)
		if err != nil {
			log.Error("msgTx.Deserialize fail")
			txRetItem = wallet_api.RawTransactionReturn{}
			txRetList = append(txRetList, &txRetItem)
			continue
		}
		txHash, err := c.btcClient.SendRawTransaction(&msgTx, true)
		if err != nil {
			log.Error("btcClient.SendRawTransaction fail")
			txRetItem = wallet_api.RawTransactionReturn{}
			txRetList = append(txRetList, &txRetItem)
			continue
		}
		if strings.Compare(msgTx.TxHash().String(), txHash.String()) != 0 {
			log.Error("broadcast transaction, tx hash mismatch", "local hash", msgTx.TxHash().String(), "hash from net", txHash.String(), "signedTx", req.RawTx)
			txRetItem = wallet_api.RawTransactionReturn{}
			txRetList = append(txRetList, &txRetItem)
			continue
		}
		txRetItem.TxHash = txHash.String()
		txRetItem.IsSuccess = true
		txRetItem.Message = "send transaction success"
		txRetList = append(txRetList, &txRetItem)
	}

	return &wallet_api.SendTransactionResponse{
		Code:   wallet_api.ApiReturnCode_APISUCCESS,
		Msg:    "send tx success",
		TxnRet: txRetList,
	}, nil
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
