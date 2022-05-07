package BLC

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func sendVersion(toAddress string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()
	// nodeAddress debug 刚开始在server里面用 nodeAddress := .... , 创建成局部变量，导致此处一直接收不到值
	payload := gobEncode(Version{NODE_VERSION, bestHeight, nodeAddress})
	request := append(commandToBytes(COMMAND_VERSION), payload...)
	sendData(toAddress, request)
}

func sendData(to string, data []byte) {
	fmt.Printf("[%s] 客户端向%s发送数据。\n", GetNowTime(), to)
	conn, err := net.Dial("tcp", to)
	if err != nil {
		panic("err")
	}
	defer conn.Close()

	// 附带要发送的数据
	_, err = io.Copy(conn, bytes.NewBuffer([]byte(data)))
	if err != nil {
		log.Panic(err)
	}
}

// 与主节点同步区块
func sendGetBlocks(toAddress string) {
	payload := gobEncode(GetBlocks{nodeAddress})
	request := append(commandToBytes(COMMAND_GETBLOCKS), payload...)
	sendData(toAddress, request)
}

// 主节点将自己的所有区块哈希发送给钱包节点
func sendInv(toAddress string, kind string, hashes [][]byte) {
	payload := gobEncode(Inv{nodeAddress, kind, hashes})
	request := append(commandToBytes(COMMAND_INV), payload...)
	sendData(toAddress, request)
}

func sendGetData(toAddress string, kind string, blockHash []byte) {
	payload := gobEncode(GetData{nodeAddress, kind, blockHash})
	request := append(commandToBytes(COMMAND_GETDATA), payload...)
	sendData(toAddress, request)
}

func sendBlock(toAddress string, block []byte) {
	payload := gobEncode(BlockData{nodeAddress, block})
	request := append(commandToBytes(COMMAND_BLOCK), payload...)
	sendData(toAddress, request)
}

func sendTx(toAddress string, tx *Transaction) {
	payload := gobEncode(Tx{nodeAddress, tx})
	request := append(commandToBytes(COMMAND_TX), payload...)
	sendData(toAddress, request)
}

//func sendData(to string,data []byte)  {
//
//	fmt.Println("客户端向服务器发送数据......")
//	conn, err := net.Dial("tcp", to)
//	if err != nil {
//		panic("error")
//	}
//	defer conn.Close()
//	// 附带要发送的数据
//	_, err = io.Copy(conn, bytes.NewReader(data))
//	if err != nil {
//		log.Panic(err)
//	}
//}
