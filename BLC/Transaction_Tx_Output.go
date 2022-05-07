package BLC

import "bytes"

type TxOutput struct {
	Value           int64
	Ripemd160Sha256 []byte // 公钥的hash
}

func (txOutput *TxOutput) Lock(address string) {
	publicKeyHash := Base58Decode([]byte(address))
	txOutput.Ripemd160Sha256 = publicKeyHash[1 : len(publicKeyHash)-4]
}

func NewTxOutput(value int64, address string) *TxOutput {
	txOutput := &TxOutput{value, nil}
	// 设置Ripemd160Sha256
	txOutput.Lock(address)
	return txOutput
}

// 判断当前消费是谁的钱
func (txOutput *TxOutput) UnLockScriptPubKeyWithAddress(address string) bool {
	publicKeyHash := Base58Decode([]byte(address))
	hash160 := publicKeyHash[1 : len(publicKeyHash)-4]
	return bytes.Compare(txOutput.Ripemd160Sha256, hash160) == 0
}
