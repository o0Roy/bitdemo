package BLC

import (
	"github.com/boltdb/bolt"
	"log"
)

// 遍历到第几个 block 就第几个的信息
type BlockchainIterator struct {
	CurrentHash []byte // 当前区块的Hash
	DB          *bolt.DB
}

// 下一个区块
func (blockchainIterator *BlockchainIterator) Next() *Block {
	var block *Block
	err := blockchainIterator.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCK_TABLE_NAME))
		if b != nil {
			currentBlockBytes := b.Get(blockchainIterator.CurrentHash)
			// 获取到当前迭代器 currentHash 所对应的区块
			block = DeserializeBlock(currentBlockBytes)

			// 更新迭代器里面的 currentHash
			blockchainIterator.CurrentHash = block.PrevBlockHash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return block
}
