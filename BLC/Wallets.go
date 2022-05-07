package BLC

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const WALLET_FILE_NAME = "wallets_%s.dat"

type Wallets struct {
	WalletsMap map[string]*Wallet
}

// 创建新钱包
func NewWallets(nodeID string) (*Wallets, error) {
	WALLET_FILE_NAME := fmt.Sprintf(WALLET_FILE_NAME, nodeID)
	if _, err := os.Stat(WALLET_FILE_NAME); os.IsNotExist(err) {
		wallets := &Wallets{}
		wallets.WalletsMap = make(map[string]*Wallet)
		return wallets, err
	}

	fileContent, err := ioutil.ReadFile(WALLET_FILE_NAME)
	if err != nil {
		log.Panic(err)
	}
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	// 钱包信息取出来
	return &wallets, err
}

// 创建一个钱包
func (w *Wallets) CreateNewWallets(nodeID string) {
	wallet := NewWallet()
	fmt.Print("new address: ")
	fmt.Println(string(wallet.GetAddress()))
	w.WalletsMap[string(wallet.GetAddress())] = wallet
	w.SaveWallets(nodeID)
}

// 将钱包信息写入到文件
func (w *Wallets) SaveWallets(nodeID string) {
	WALLET_FILE_NAME := fmt.Sprintf(WALLET_FILE_NAME, nodeID)
	var content bytes.Buffer

	// 注册的目的：为了可以序列化任何数据
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(&w)
	if err != nil {
		log.Panic(err)
	}
	// 将序列化厚的数据写如文件，会覆盖原文件。
	err = ioutil.WriteFile(WALLET_FILE_NAME, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
