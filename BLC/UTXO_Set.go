package BLC

import (
	"bytes"
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

// 遍历数据库，读取所有的未花费的UTXO，然后将所有的UTXO存储到数据库
// reset
// 去遍历数据库
// 【】*TXoutputs

//[string]*TXOutputs
//
//txHash, TXOutputs := range txOutputsMap {
//
//}

const UTXO_TALBE_NAME = "utxoTableName"

type UTXOSet struct {
	Blockchain *Blockchain
}

// 重置 utxo 表
func (utxoSet *UTXOSet) ResetUTXOSet() {
	err := utxoSet.Blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXO_TALBE_NAME))
		if b != nil {
			err := tx.DeleteBucket([]byte(UTXO_TALBE_NAME))
			if err != nil {
				log.Println(err)
			}
		}
		b, _ = tx.CreateBucket([]byte(UTXO_TALBE_NAME))
		if b != nil {
			txOutputsMap := utxoSet.Blockchain.FindUTXOMap()
			//fmt.Println(txOutputsMap)
			for keyHash, outs := range txOutputsMap {
				//UTXO_TALBE_NAME 表按交易的 txHash, utxo 类型存储
				txHash, _ := hex.DecodeString(keyHash) // keyHash 转为 []byte 类型
				b.Put(txHash, outs.Serialize())
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (utxoSet *UTXOSet) findUTXOForAddress(address string) []*UTXO {
	var utxos []*UTXO
	utxoSet.Blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXO_TALBE_NAME))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("Key=%s, value=%v\n", k, v)
			txOutputs := DeserializeBlockTxOutputs(v)
			for _, utxo := range txOutputs.UTXOS {
				if utxo.Output.UnLockScriptPubKeyWithAddress(address) {
					utxos = append(utxos, utxo)
				}
			}
		}
		return nil
	})
	return utxos
}

func (utxoSet *UTXOSet) GetBalance(address string) int64 {
	UTXOS := utxoSet.findUTXOForAddress(address)
	var amount int64
	for _, utxo := range UTXOS {
		amount += utxo.Output.Value
	}
	return amount
}

// 返回还未打包的 UTXO
func (utxoSet *UTXOSet) FindUnPackageSpendableUTXOS(from string, txs []*Transaction) []*UTXO {
	var unUTXOs []*UTXO
	spentTxOutputs := make(map[string][]int) // {hash:[0, 2, 3]}

	for _, tx := range txs {
		// Vin
		if !tx.IsCoinbaseTransaction() { // 不是创世区块
			for _, in := range tx.Vins {
				publicKeyHash := Base58Decode([]byte(from))
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
			if out.UnLockScriptPubKeyWithAddress(from) {
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
	return unUTXOs
}

func (utxoSet *UTXOSet) FindSpendableUTXOS(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {
	unPackageUTXOS := utxoSet.FindUnPackageSpendableUTXOS(from, txs)
	spentableUTXO := make(map[string][]int)
	var money int64 = 0
	for _, UTXO := range unPackageUTXOS {
		money += UTXO.Output.Value
		txHash := hex.EncodeToString(UTXO.TxHash)
		spentableUTXO[txHash] = append(spentableUTXO[txHash], UTXO.Index)
		if money >= amount {
			return money, spentableUTXO
		}
		//map[fkasjfkasjlfkjlwejlf] = [0, 1] 返回某个人某一笔 out 的坐标集
	}
	utxoSet.Blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXO_TALBE_NAME))
		if b != nil {
			c := b.Cursor()
		UTXOBREAK:
			for k, v := c.First(); k != nil; k, v = c.Next() {
				txOutputs := DeserializeBlockTxOutputs(v)
				for _, utxo := range txOutputs.UTXOS {
					money += utxo.Output.Value
					txHash := hex.EncodeToString(utxo.TxHash)
					spentableUTXO[txHash] = append(spentableUTXO[txHash], utxo.Index)
					if money >= amount {
						break UTXOBREAK
					}
				}
			}
		}
		return nil
	})
	if money < amount {
		log.Panic("余额不足")
	}
	return money, spentableUTXO
}

// 根据最新一个区块更新 UTXO 表
func (utxoSet *UTXOSet) Update() {

	// 最新的 Block
	block := utxoSet.Blockchain.Iterator().Next()

	ins := []*TxInput{} // 要删除的交易记录
	outsMap := make(map[string]*TxOutputs)
	// 找到要删除的 Input
	for _, tx := range block.Txs {
		for _, in := range tx.Vins {
			ins = append(ins, in)
		}
	}

	// 本区块内的 out 和 in 自我消费筛查
	for _, tx := range block.Txs {
		utxos := []*UTXO{}
		for index, out := range tx.Vouts {
			isSpent := false
			for _, in := range ins {
				// 比较交易的 index ，比较交易的 TxHash ，比较交易的公钥
				if in.Vout == index && bytes.Compare(tx.TxHash, in.TxHash) == 0 && bytes.Compare(out.Ripemd160Sha256, Ripemd160Sha256(in.PublicKey)) == 0 {
					isSpent = true
					continue
				}
			}
			if isSpent == false {
				utxo := &UTXO{tx.TxHash, index, out}
				utxos = append(utxos, utxo)
			}
		}
		if len(utxos) > 0 {
			txHash := hex.EncodeToString(tx.TxHash)
			outsMap[txHash] = &TxOutputs{utxos}
		}
	}

	err := utxoSet.Blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UTXO_TALBE_NAME))
		if b != nil {
			for _, in := range ins {
				txOutputsBytes := b.Get(in.TxHash)
				if len(txOutputsBytes) == 0 { // 该笔交易不在之前记录中
					continue
				}
				txOutputs := DeserializeBlockTxOutputs(txOutputsBytes)
				UTXOS := []*UTXO{}
				// 判断是否需要删除
				isDelete := false
				for _, utxo := range txOutputs.UTXOS {
					if in.Vout == utxo.Index && bytes.Compare(utxo.Output.Ripemd160Sha256, Ripemd160Sha256(in.TxHash)) == 0 {
						isDelete = true
					} else {
						UTXOS = append(UTXOS, utxo)
					}
				}
				if isDelete {
					b.Delete(in.TxHash) // 删除某个 key
					if len(UTXOS) > 0 {
						preTxOutputs := outsMap[hex.EncodeToString(in.TxHash)]
						preTxOutputs.UTXOS = append(preTxOutputs.UTXOS, UTXOS...)
						outsMap[hex.EncodeToString(in.TxHash)] = preTxOutputs
					}
				}
			}
			// 新增
			for keyHash, outPuts := range outsMap {
				keyHashBytes, _ := hex.DecodeString(keyHash)
				b.Put(keyHashBytes, outPuts.Serialize())
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
