package BLC

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"
)

const DB_NAME = "blockchain_%s.db" // 数据库名字
const BLOCK_TABLE_NAME = "blocks"  // 表的名字
const GENESIS_MONEY = 100

type Blockchain struct {
	Tip []byte // 最新区块的 Hash
	DB  *bolt.DB
}

// 迭代器
func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{blockchain.Tip, blockchain.DB}
}

// 1. 创建带有创世区块的区块链
func CreateBlockWithGenesisBlock(address string, nodeID string) *Blockchain {
	// 格式化数据库名称
	DB_NAME := fmt.Sprintf(DB_NAME, nodeID)
	// 判断数据库是否存在
	if DBExists(DB_NAME) {
		fmt.Println("创世区块已经存在。")
		os.Exit(1)
	}
	var genesisHash []byte
	fmt.Println("创世区块创建完成。")
	// 创建创世区块的时候要尝试创建或打开数据库
	db, err := bolt.Open(DB_NAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		// 创建数据库表
		b, err := tx.CreateBucketIfNotExists([]byte(BLOCK_TABLE_NAME))
		if err != nil {
			log.Panic(err)
		}
		// 创建创世区块
		txCoinbase := NewConbaseTransaction(address)
		genesisBlock := CreateGenesisBlock([]*Transaction{txCoinbase})
		// 将创世区块存储到表当中
		err = b.Put(genesisBlock.Hash, genesisBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), genesisBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		genesisHash = genesisBlock.Hash
		return nil
	})
	return &Blockchain{genesisHash, db}
}

// 2. 增加区块到区块链当中
func (blc *Blockchain) AddBlockToBlockchain(txs []*Transaction) {
	// 创建新区块
	err := blc.DB.Update(func(tx *bolt.Tx) error {
		// 1. 获取表
		b := tx.Bucket([]byte(BLOCK_TABLE_NAME))
		if b != nil {
			// 获取最新区块
			blockBytes := b.Get(blc.Tip)
			block := DeserializeBlock(blockBytes)
			// 2. 创建新区块
			newBlock := NewBlock(txs, block.Height+1, block.Hash)
			// 3. 将区块序列化并且存储到数据库当中
			err := b.Put(newBlock.Hash, newBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			// 4. 更新数据库中 l 的对应 Hash
			err = b.Put([]byte("l"), newBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
			// 5. 更新 blockchain 的 Tip
			blc.Tip = newBlock.Hash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// 遍历输出所有区块的信息
func (blc *Blockchain) PrintChain() {

	blockchainIterator := blc.Iterator()

	for {
		block := blockchainIterator.Next()
		fmt.Printf("Height:%d\n", block.Height)
		fmt.Printf("preBlockHash:%x\n", block.PrevBlockHash)
		//fmt.Printf("Txs:%v\n", block.Txs)
		fmt.Printf("Timestamp:%v\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash:%x.\n", block.Hash)
		fmt.Printf("nonce:%d\n", block.Nonce)
		fmt.Println("Txs:")
		for _, tx := range block.Txs {
			fmt.Printf("\tTxHash:  %x\n", tx.TxHash)
			fmt.Println("\tVins:")
			for _, in := range tx.Vins {
				fmt.Printf("\t\tTxInHash:    %x\n", in.TxHash)
				fmt.Printf("\t\tVout:      %d\n", in.Vout)
				fmt.Printf("\t\tScriptSin: %x\n", in.PublicKey)
			}
			fmt.Println("\tVouts:")
			for _, out := range tx.Vouts {
				fmt.Printf("\t\tValue:        %d\n", out.Value)
				fmt.Printf("\t\tScriptPubkey: %x\n", out.Ripemd160Sha256)
			}
		}
		fmt.Println()
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
}

// 判断数据库是否存在
func DBExists(DB_NAME string) bool {

	if _, err := os.Stat(DB_NAME); os.IsNotExist(err) {
		return false
	}
	return true
}

// 获取 BLockchain Object
func BlockchainObject(nodeID string) *Blockchain {
	DB_NAME := fmt.Sprintf(DB_NAME, nodeID)
	// 判断数据库是否存在
	if DBExists(DB_NAME) == false {
		fmt.Println("数据库不存在！")
		os.Exit(1)
	}

	db, err := bolt.Open(DB_NAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	var tip []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCK_TABLE_NAME))
		if b != nil {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	return &Blockchain{tip, db}
}

// 返回对应地址的 TxOutput
func (blockchain *Blockchain) UnUTXOs(address string, txs []*Transaction) []*UTXO {

	var unUTXOs []*UTXO
	spentTxOutputs := make(map[string][]int) // {hash:[0, 2, 3]}
	for _, tx := range txs {
		// Vin
		if tx.IsCoinbaseTransaction() == false { // 不是创世区块
			for _, in := range tx.Vins {
				publicKeyHash := Base58Decode([]byte(address))
				ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]
				if in.UnLockRipemd160Sha256(ripemd160Hash) { // 是否能够解锁
					key := hex.EncodeToString(in.TxHash)
					//fmt.Println(key)
					spentTxOutputs[key] = append(spentTxOutputs[key], in.Vout)
				}
			}
		}
	}
	for _, tx := range txs {
	jump_1:
		for index, out := range tx.Vouts {
			if out.UnLockScriptPubKeyWithAddress(address) {
				if len(spentTxOutputs) == 0 {
					utxo := &UTXO{tx.TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash, indexArray := range spentTxOutputs {
						txHashStr := hex.EncodeToString(tx.TxHash)
						if hash == txHashStr {
							var isUnSpentUTXO bool
							for _, outIndex := range indexArray {
								if index == outIndex {
									isUnSpentUTXO = true
									continue jump_1
								}
							}
							if !isUnSpentUTXO {
								utxo := &UTXO{tx.TxHash, index, out}
								unUTXOs = append(unUTXOs, utxo)
							}
						} else {
							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}
	}

	blockIterator := blockchain.Iterator()

	for {
		block := blockIterator.Next()

		//tt := hex.EncodeToString(block.Txs[0].TxHash)
		//fmt.Println(tt)
		for i := len(block.Txs) - 1; i >= 0; i-- {
			// 从数据库取出来的 Transaction 要从最新一条 Transaction 开始往回回溯
			// 倒序回溯才不会出错。
			tx := block.Txs[i]

			//for _, tt := range tx.Vins {
			//	fmt.Println(tt)
			//}
			//fmt.Println(">>>>>>>>>>>>>>>>>>>>>")
			//for _, tt := range tx.Vouts {
			//	fmt.Println(tt)
			//}
			//fmt.Println()

			// Vins
			if !tx.IsCoinbaseTransaction() { // 不是创世区块
				for _, in := range tx.Vins {
					publicKeyHash := Base58Decode([]byte(address))
					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]
					if in.UnLockRipemd160Sha256(ripemd160Hash) { // 是否能够解锁
						key := hex.EncodeToString(in.TxHash)
						//fmt.Println(key)
						spentTxOutputs[key] = append(spentTxOutputs[key], in.Vout)
					}
				}
			}

			// Vouts
		work:
			for index, out := range tx.Vouts {
				if out.UnLockScriptPubKeyWithAddress(address) {
					if spentTxOutputs != nil {
						if len(spentTxOutputs) != 0 {
							var isSpentUTXO bool // 默认为 false
							for txHash, indexArray := range spentTxOutputs {
								//fmt.Println(spentTxOutputs)
								for _, i := range indexArray {
									key := hex.EncodeToString(tx.TxHash)
									if index == i && txHash == key {
										isSpentUTXO = true
										continue work // 该笔 out 相同，则进入下一笔 out 对比。
									}
								}
							}
							if !isSpentUTXO {
								utxo := &UTXO{tx.TxHash, index, out}
								unUTXOs = append(unUTXOs, utxo)
							}
						} else {
							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	//fmt.Println(spentTxOutputs)
	//fmt.Println(unUTXOs)
	return unUTXOs
}

func (blockchain *Blockchain) MineNewBlock(from []string, to []string, amount []string, nodeID string) *Blockchain {
	// 1. 通过相关算法建立 Transaction 数组

	var utxoSet = &UTXOSet{blockchain}
	var txs []*Transaction

	for index, address := range from {
		value, _ := strconv.Atoi(amount[index])
		tx := NewSimpleTransaction(address, to[index], int64(value), utxoSet, txs, nodeID) // 把当前区块的交易 txs 传入
		txs = append(txs, tx)
	}

	// 奖励
	tx := NewConbaseTransaction(from[0])
	txs = append(txs, tx)

	var block *Block
	blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCK_TABLE_NAME))
		if b != nil {
			hash := b.Get([]byte("l"))
			blockBytes := b.Get([]byte(hash))
			block = DeserializeBlock(blockBytes)
		}
		return nil
	})

	// 建立新区块之前对 txs 进行签名验证
	_txs := []*Transaction{}
	for _, tx := range txs {
		if blockchain.VerifyTransaction(tx, txs) == false {
			log.Panic("无签名 - 非法交易。")
		}
		_txs = append(_txs, tx) //
	}

	// 2. 建立新的区块
	block = NewBlock(txs, block.Height+1, block.Hash)
	fmt.Printf("新区块 Hash：\r%x\n", block.Hash)
	// 将新区块存储到数据库当中
	blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCK_TABLE_NAME))
		if b != nil {
			b.Put(block.Hash, block.Serialize())
			b.Put([]byte("l"), block.Hash)
			blockchain.Tip = block.Hash
		}
		return nil
	})
	return nil
}

// 查询余额
func (blockchain *Blockchain) GetBalance(address string) int64 {
	utxos := blockchain.UnUTXOs(address, []*Transaction{})

	var amount int64
	for _, utxo := range utxos {
		//fmt.Println(utxo.Output)
		amount = amount + utxo.Output.Value
	}
	return amount
}

// 查找可以用的UTXO
func (blockchain *Blockchain) FindSpendAbleUTXOs(from string, amount int, txs []*Transaction) (int64, map[string][]int) {
	// 1. 获取所有UTXO
	utxos := blockchain.UnUTXOs(from, txs)

	spendAbleUTXO := make(map[string][]int)
	var value int64
	// 2. 遍历utxo
	for _, utxo := range utxos {
		value = value + utxo.Output.Value
		hash := hex.EncodeToString(utxo.TxHash)
		spendAbleUTXO[hash] = append(spendAbleUTXO[hash], utxo.Index)
		if value >= int64(amount) {
			break
		}
	}
	if value < int64(amount) {
		fmt.Printf("%s has insufficient balance.", from)
		os.Exit(1)
	}
	return value, spendAbleUTXO
}

func (blockchain *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey, txs []*Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}
	prevTXs := make(map[string]Transaction) // 找出所有 vin 的出处 output
	for _, vin := range tx.Vins {
		prevTX, err := blockchain.FindTransaction(vin.TxHash, txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}

func (blockchain *Blockchain) FindTransaction(ID []byte, txs []*Transaction) (Transaction, error) {
	for _, tx := range txs {
		if bytes.Compare(tx.TxHash, ID) == 0 {
			return *tx, nil
		}
	}
	bci := blockchain.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Txs {
			if bytes.Compare(tx.TxHash, ID) == 0 {
				return *tx, nil
			}
		}
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		// 判断是否找到创世区块
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	return Transaction{}, nil
}

// 验证数字签名
func (blockchain *Blockchain) VerifyTransaction(tx *Transaction, txs []*Transaction) bool {

	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vins {
		prevTX, err := blockchain.FindTransaction(vin.TxHash, txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}
	return tx.Verify(prevTXs)
}

// 查找 UTXOMap
func (blockchain *Blockchain) FindUTXOMap() map[string]*TxOutputs {

	blcIterator := blockchain.Iterator()
	// a := make(map[string]map[string][]int) //另一种方法

	spentUTXOsMap := make(map[string][]*TxInput) //存储已经花费的 UTXO 的信息
	utxoMaps := make(map[string]*TxOutputs)
	for {
		block := blcIterator.Next()
		for i := len(block.Txs) - 1; i >= 0; i-- {
			txOutputs := &TxOutputs{[]*UTXO{}}
			tx := block.Txs[i]
			// coinbase

			if tx.IsCoinbaseTransaction() == false {
				for _, txInput := range tx.Vins {
					txHash := hex.EncodeToString(txInput.TxHash)
					spentUTXOsMap[txHash] = append(spentUTXOsMap[txHash], txInput)
				}
			}
			txHash := hex.EncodeToString(tx.TxHash) // 当前交易的 txHash
			txInputs := spentUTXOsMap[txHash]
			if len(txInputs) > 0 {
			WorkOutLoop:
				for index, out := range tx.Vouts {
					for _, in := range txInputs {
						outPublicKey := out.Ripemd160Sha256
						inPublicKey := in.PublicKey
						if bytes.Compare(outPublicKey, Ripemd160Sha256(inPublicKey)) == 0 {
							if index == in.Vout {
								continue WorkOutLoop
							} else {
								utxo := &UTXO{tx.TxHash, index, out}
								txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
							}
						}
					}
				}
			} else {
				for index, out := range tx.Vouts {
					utxo := &UTXO{tx.TxHash, index, out}
					txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
				}
			}
			utxoMaps[txHash] = txOutputs
		}
		//workOutLoop:
		//	for index, out := range tx.Vouts {
		//		txInputs := spentUTXOsMap[txHash] // 取出花费表中交易哈希为 txHash 的消费记录
		//		if len(txInputs) > 0 {            // 如果记录为 0 ，则未被消费，直接加到 UTXO
		//			for _, in := range txInputs { // 判断花了几笔
		//				outPublicKey := out.Ripemd160Sha256
		//				inPublicKey := in.PublicKey
		//				if bytes.Compare(outPublicKey, Ripemd160Sha256(inPublicKey)) == 0 { // 判断是否为同一个人交易
		//					if index == in.Vout { // 判断花费的是否为第 index 的 out
		//						continue workOutLoop
		//					}
		//				}
		//			}
		//			utxo := &UTXO{tx.TxHash, index, out}
		//			txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
		//		} else {
		//			utxo := &UTXO{tx.TxHash, index, out}
		//			txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
		//		}
		//	}
		//	utxoMaps[txHash] = txOutputs
		//}
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		// 找到创世区块退出
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return utxoMaps
}

func (bc *Blockchain) GetBestHeight() int64 {
	block := bc.Iterator().Next()
	return block.Height
}

func (bc *Blockchain) GetAllBlockHashes() [][]byte {
	blockIterator := bc.Iterator()
	var blockHashes [][]byte
	for {
		block := blockIterator.Next()
		blockHashes = append(blockHashes, block.Hash)
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		// 找到创世区块退出
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return blockHashes
}

// 根据区块 hash 获取区块
func (bc *Blockchain) GetBlock(blockHash []byte) ([]byte, error) {

	var blockBytes []byte
	err := bc.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(BLOCK_TABLE_NAME))
		if b != nil {
			blockBytes = b.Get(blockHash)
		}
		return nil
	})
	return blockBytes, err
}

// 同步后存储区块
func (bc *Blockchain) AddBlock(block *Block) error {
	err := bc.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(BLOCK_TABLE_NAME))
		if b != nil {
			blockExit := b.Get(block.Hash) // 判断该区块是否存在
			if blockExit != nil {
				// 如果存在，则不处理
				return nil
			}
			err := b.Put(block.Hash, block.Serialize())
			if err != nil {
				log.Panic(err)
			}
			// 当前数据库中最新区块的 hash
			blockHash := b.Get([]byte("l"))
			blockBytes := b.Get(blockHash)
			blockInDB := DeserializeBlock(blockBytes)
			if block.Height > blockInDB.Height {
				b.Put([]byte("l"), block.Hash)
				bc.Tip = block.Hash
			}
		}
		return nil
	})
	return err
}
