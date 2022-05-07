package BLC

// 创建创世区块
func (cli *CLI) resetUTXOSet(nodeID string) {
	blockchain := BlockchainObject(nodeID)
	defer blockchain.DB.Close()
	utxoSet := &UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()
	//fmt.Println(utxoMap["05e9384019d94f555a1aeb4ad435294b2aac3742a2722ec43a823c700856ff6b"])
	//fmt.Println(utxoMap["cc43ea4d82a8a23c408268947bd89ce01664f14da41e71db8fe5502ab85ee3ce"])
	//fmt.Println(utxoMap)
}
