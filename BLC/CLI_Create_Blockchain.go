package BLC

// 创建创世区块
func (cli *CLI) createGenesisBlockchain(address string, nodeID string) {
	blockchain := CreateBlockWithGenesisBlock(address, nodeID)
	defer blockchain.DB.Close()
	utxoSet := &UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()
}
