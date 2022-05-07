package BLC

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

func handleVersion(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload Version
	dataBytes := request[COMMANDLENGTH:] // 截取命令后面的数据
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// 获取当前区块高度
	bestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight
	if bestHeight > foreignerBestHeight {
		// 自己比外部节点高就将自己的的高度发送给外部节点，外部节点再来同步
		sendVersion(payload.AddrFrom, bc)
	} else if bestHeight < foreignerBestHeight {
		// 向主节点要信息
		sendGetBlocks(payload.AddrFrom)
	}
	if !nodeIsKnown(payload.AddrFrom) {
		knowNodes = append(knowNodes, payload.AddrFrom)
	}
}

func handleAddr(request []byte, bc *Blockchain) {

}

func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload BlockData

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockBytes := payload.Block
	block := DeserializeBlock(blockBytes)
	fmt.Printf("[%s] Received a new block:%x\n", GetNowTime(), block.Hash)
	// 添加区块到本地数据库
	bc.AddBlock(block)
	UTXOSet := &UTXOSet{bc}
	UTXOSet.Update()
	fmt.Printf("[%s] Added block:%x\n", GetNowTime(), block.Hash)
	fmt.Printf("[%s] 本地数据库重置。", GetNowTime())
	if len(transactionArray) > 0 {
		sendGetData(payload.AddrFrom, BLOCK_TYPE, transactionArray[0])
		transactionArray = transactionArray[1:]
	} else {

		//UTXOSet := &UTXOSet{bc}
		//UTXOSet.ResetUTXOSet()
	}

}

func handleGetblocks(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload GetBlocks

	dataBytes := request[COMMANDLENGTH:]
	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blocks := bc.GetAllBlockHashes()
	sendInv(payload.AddrFrom, BLOCK_TYPE, blocks)
}

func handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload GetData

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == BLOCK_TYPE {
		block, err := bc.GetBlock([]byte(payload.Hash))
		if err != nil {
			return
		}
		sendBlock(payload.AddrFrom, block)
	}
	if payload.Type == TX_TYPE {
		tx := memoryTxPool[hex.EncodeToString(payload.Hash)]
		sendTx(payload.AddrFrom, tx)
	}
}

func handleTx(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload Tx
	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	tx := payload.Tx
	memoryTxPool[hex.EncodeToString(tx.TxHash)] = tx

	if nodeAddress == knowNodes[0] {
		// 给矿工节点发送交易 hash
		for _, nodeAddr := range knowNodes {
			if nodeAddr != nodeAddress && nodeAddr != payload.AddrFrom {
				sendInv(nodeAddr, TX_TYPE, [][]byte{tx.TxHash})
			}
		}
	}

	// 矿工进行挖矿验证
	if len(minerAddress) > 0 {
		utxoSet := &UTXOSet{bc}
		txs := []*Transaction{tx}
		// 奖励
		coinBaseTx := NewConbaseTransaction(minerAddress)
		txs = append(txs, coinBaseTx)

		_txs := []*Transaction{}
		// fmt.Printf("[%s] 开始数字签名验证。\n", GetNowTime())
		for index, tx := range txs {
			fmt.Printf("[%s] 开始第%d次验证。\n", GetNowTime(), index)

			// 数字签名失败
			if bc.VerifyTransaction(tx, _txs) != true {
				log.Panic("Error: Invalid transaction")
			}
			fmt.Printf("[%s] 第%s次验证成功\n", GetNowTime(), index)
			_txs = append(_txs, tx)
		}
		fmt.Printf("[%s] 数字签名成功。", GetNowTime())

		// 1.通过相关算法建立 Transaction 数组
		var block *Block
		bc.DB.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(BLOCK_TABLE_NAME))
			if b != nil {
				hash := b.Get([]byte("l"))
				blockBytes := b.Get(hash)
				block = DeserializeBlock(blockBytes)
			}
			return nil
		})

		// 2. 建立新的区块
		block = NewBlock(txs, block.Height+1, block.Hash)

		// 将新区块存储到数据库
		bc.DB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(BLOCK_TABLE_NAME))
			if b != nil {
				b.Put(block.Hash, block.Serialize())
				b.Put([]byte("l"), block.Hash)
				bc.Tip = block.Hash
			}
			return nil
		})
		utxoSet.Update()
		sendBlock(knowNodes[0], block.Serialize())
	}
}

func handleInv(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload Inv

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == BLOCK_TYPE {
		transactionArray = payload.Items
		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, BLOCK_TYPE, blockHash)
		if len(payload.Items) >= 1 {
			transactionArray = payload.Items[1:]
		}
	}
	if payload.Type == TX_TYPE {
		txHash := payload.Items[0]
		if memoryTxPool[hex.EncodeToString(txHash)] == nil {
			sendGetData(payload.AddrFrom, TX_TYPE, txHash)
		}
	}
}
