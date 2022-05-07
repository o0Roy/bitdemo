package BLC

import "fmt"

// 查询余额
func (cli *CLI) getBalance(address string, nodeID string) {
	fmt.Println("地址：" + address)
	blockchain := BlockchainObject(nodeID)
	defer blockchain.DB.Close()
	utxoSet := &UTXOSet{blockchain}
	amount := utxoSet.GetBalance(address)
	//amount := blockchain.GetBalance(address)
	fmt.Printf("%s一共有%d个Token\n", address, amount)
}
