package BLC

import (
	"fmt"
	"os"
)

// 创建创世区块
func (cli *CLI) startNode(nodeID string, minerAdd string) {
	if minerAdd == "" || IsValidForAddress([]byte(minerAdd)) {
		// 启动服务器
		fmt.Printf("[%s] 启动服务器：localhost：%s\n", GetNowTime(), nodeID)
		startServer(nodeID, minerAdd)
	} else {
		fmt.Println("指定地址无效。")
		os.Exit(0)
	}
}
