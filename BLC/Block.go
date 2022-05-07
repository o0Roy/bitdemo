package BLC

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Height        int64          // 1.区块高度
	PrevBlockHash []byte         // 2.上一个区块HASH
	Txs           []*Transaction // 3.交易数据
	Timestamp     int64          // 4.时间戳
	Hash          []byte         // 5.Hash
	Nonce         int64          // 6.Nonce
}

// 1.创建新的区块
func NewBlock(txs []*Transaction, height int64, preBlockHash []byte) *Block {
	// 创建区块
	block := &Block{height, preBlockHash, txs, time.Now().Unix(), nil, 0}

	// 调用工作 proof of work方法并且返回 Hash 和 Nonce
	pow := NewProofOfWork(block)

	// hash 前面有000000
	hash, nonce := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// 2.生成创世区块
func CreateGenesisBlock(txs []*Transaction) *Block {
	return NewBlock(txs, 1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

// 返回 TXS 转换 []byte
func (block *Block) HashTransactions() []byte {
	//var txHashes [][]byte
	//var txHash [32]byte
	//for _, tx := range block.Txs {
	//	txHashes = append(txHashes, tx.TxHash)
	//}
	//txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	//return txHash[:]
	var transactions [][]byte
	for _, tx := range block.Txs {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)
	return mTree.RootNode.Data

}

// 将区块序列号成字节数组
func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// 反序列化
func DeserializeBlock(blockBytes []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
