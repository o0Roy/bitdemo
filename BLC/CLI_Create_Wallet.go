package BLC

import "fmt"

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := NewWallets(nodeID)
	wallets.CreateNewWallets(nodeID)
	// 把所有数据存储

	//fmt.Print("address:")
	fmt.Println(len(wallets.WalletsMap))
}
