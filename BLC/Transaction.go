package BLC

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"math/big"
	"time"
)

type Transaction struct {
	TxHash []byte      // 交易哈希
	Vins   []*TxInput  // 输入
	Vouts  []*TxOutput // 输出
}

// Hash 序列化
func (tx *Transaction) HashTransaction() {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	resultBytes := bytes.Join([][]byte{IntToHex(time.Now().Unix()), result.Bytes()}, []byte{})
	hash := sha256.Sum256(resultBytes)
	tx.TxHash = hash[:]
}

// 1.交易的创建分两种情况
// 1.1 创世区块创建时的Transaction
func NewConbaseTransaction(address string) *Transaction {
	// 代表消费							 Vout 为负责消费的 output
	txInput := &TxInput{[]byte{}, -1, nil, []byte{}}
	// 输出交易
	txOutput := NewTxOutput(10, address)

	txConbase := &Transaction{[]byte{}, []*TxInput{txInput}, []*TxOutput{txOutput}}
	// 设置 Txhash
	txConbase.HashTransaction()
	return txConbase
}

// 1.2 转账时产生的Transaction
func NewSimpleTransaction(from string, to string, amount int64, utxoSet *UTXOSet, txs []*Transaction, nodeID string) *Transaction {

	// 1.有一个函数，返回 from 这个人的所有未花费交易输出所对应的 Transaction
	//unSpentTx := UnUTXOs(from)
	//fmt.Println(unSpentTx)
	wallets, _ := NewWallets(nodeID)
	wallet := wallets.WalletsMap[from]
	// 2.通过一个函数，返回
	money, spendAbleUTXODic := utxoSet.FindSpendableUTXOS(from, amount, txs)
	var txInputs []*TxInput
	var txOutputs []*TxOutput
	//str, _ := hex.DecodeString("b59408e1af35f1e489a6a44d7f3c562e0536696c113ad3be6674628ef580227b")
	// 代表消费							 Vout 为负责消费的 output
	for txHash, indexArray := range spendAbleUTXODic {
		txHashBytes, _ := hex.DecodeString(txHash)
		for _, index := range indexArray {
			txInput := &TxInput{txHashBytes, index, nil, wallet.Publickey}
			txInputs = append(txInputs, txInput)
		}
	}

	// 转账
	txOutput := NewTxOutput(int64(amount), to)
	txOutputs = append(txOutputs, txOutput)
	// 找零
	txOutput = NewTxOutput(int64(money)-int64(amount), from)
	txOutputs = append(txOutputs, txOutput)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	// 设置 Txhash
	tx.HashTransaction()

	// 进行签名
	utxoSet.Blockchain.SignTransaction(tx, wallet.Privatekey, txs)
	return tx
}

// 序列化
func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	txCopy := tx
	txCopy.TxHash = []byte{}
	hash := sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}

	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}
	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Sha256
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[inID].PublicKey = nil

		// 签名代码
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.TxHash)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vins[inID].Signature = signature
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []*TxInput
	var outputs []*TxOutput
	for _, vin := range tx.Vins {
		inputs = append(inputs, &TxInput{vin.TxHash, vin.Vout, nil, nil})
	}
	for _, vout := range tx.Vouts {
		outputs = append(outputs, &TxOutput{vout.Value, vout.Ripemd160Sha256})
	}
	txCopy := Transaction{tx.TxHash, inputs, outputs}
	return txCopy
}

func (tx *Transaction) IsCoinbaseTransaction() bool {
	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}

// 数字签名验证
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbaseTransaction() {
		return true
	}
	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("ERROR: Previous transaction is not correct.")
		}
	}
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	for inID, vin := range tx.Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Sha256
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[inID].PublicKey = nil

		// 私钥 ID
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.PublicKey[(keyLen / 2):])
		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.TxHash, &r, &s) == false {
			return false
		}
	}
	return true
}
