package BLC

// 输出所有区块
func (cli *CLI) printChain(nodeID string) {

	blockchain := BlockchainObject(nodeID)
	defer blockchain.DB.Close()
	blockchain.PrintChain()
}
