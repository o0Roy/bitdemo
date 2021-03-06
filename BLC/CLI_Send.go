package BLC

import (
	"fmt"
	"strconv"
)

// 转账
func (cli *CLI) send(from []string, to []string, amount []string, nodeID string, mineNow bool) {
	blockchain := BlockchainObject(nodeID)
	defer blockchain.DB.Close()
	utxoSet := &UTXOSet{blockchain}
	if mineNow {
		blockchain.MineNewBlock(from, to, amount, nodeID)
		// 转账成功后，需要更新一下
		//utxoSet.Update()
		utxoSet.ResetUTXOSet()
	} else {
		// 把交易发送到旷工节点去验证
		fmt.Println("由矿工节点处理。")
		value, _ := strconv.Atoi(amount[0])
		tx := NewSimpleTransaction(from[0], to[0], int64(value), utxoSet, []*Transaction{}, nodeID)
		sendTx(knowNodes[0], tx)
	}

}
