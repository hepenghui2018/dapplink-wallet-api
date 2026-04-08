package evmbase

func IsEthTransfer(tx *Eip1559DynamicFeeTx) bool {
	if tx.ContractAddress == "" || tx.ContractAddress == NativeToken {
		return true
	}
	return false
}

type Eip1559DynamicFeeTx struct {
	ChainId              string `json:"chain_id"`
	Nonce                uint64 `json:"nonce"`
	FromAddress          string `json:"from_address"`
	ToAddress            string `json:"to_address"`
	GasLimit             uint64 `json:"gas_limit"`
	Gas                  uint64 `json:"Gas"`
	MaxFeePerGas         string `json:"max_fee_per_gas"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas"`
	Amount               string `json:"amount"`
	ContractAddress      string `json:"contract_address"`
	Signature            string `json:"signature,omitempty"`
}
