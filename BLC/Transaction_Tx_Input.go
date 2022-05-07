package BLC

import "bytes"

type TxInput struct {
	TxHash    []byte // 交易的Hash
	Vout      int    // 存储TxOutput坐标
	Signature []byte // 数字签名
	PublicKey []byte // 公钥,钱包里
}

// 判断当前消费是谁的钱
func (txInput *TxInput) UnLockRipemd160Sha256(ripemd160Sha256 []byte) bool {

	publicKey := Ripemd160Sha256(txInput.PublicKey)
	return bytes.Compare(publicKey, ripemd160Sha256) == 0
}
