package BLC

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct {
}

// 输出命令行用处
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tcreateWallet -- 创建钱包")
	fmt.Println("\tcreateBlockchain -address -- 交易数据")
	fmt.Println("\taddressList -- 输出钱包列表")
	fmt.Println("\tsend -from from -to To -amount AMOUNT -- 交易数据")
	fmt.Println("\tprintChain -- 输出区块信息.")
	fmt.Println("\tgetBalance -address -- 获取区块金额")
	fmt.Println("\tresetUTXOSet -- 重置")
	fmt.Println("\tstartnode --miner -- 启动节点")
}

func (cli *CLI) Run() {
	isValidArgs()

	// 获取节点 ID
	// 设置 ID
	// export NODE_ID=8888  // 等号左右不能有空格，直接在终端中输入，此为设置环境变量
	// 读取节点 ID

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Println("NODE_ID env. var is not set!")
		os.Exit(1)
	}
	fmt.Printf("NODE_ID:%s\n", nodeID)

	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createBlockchain", flag.ExitOnError)
	addressListCmd := flag.NewFlagSet("addressList", flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	resetUTXOSetCmd := flag.NewFlagSet("resetUTXOSet", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startNode", flag.ExitOnError)
	// 为命令添加参数
	// createBlockchain 命令添加参数
	flagCreateBlockchainWithAddress := createBlockchainCmd.String("address", "", "创建创世区块的地址。")
	// send 命令添加参数
	flagFrom := sendBlockCmd.String("from", "", "转账源地址")    // 可以在 addBlock 后面加上一个 -data
	flagTo := sendBlockCmd.String("to", "", "转账目的地址")       // 可以在 addBlock 后面加上一个 -data
	flagAmount := sendBlockCmd.String("amount", "", "转账金额") // 可以在 addBlock 后面加上一个 -data
	flagMine := sendBlockCmd.Bool("mine", false, "是否在当前节点中立即验证。")
	// getBalance 命令添加参数
	getBalanceWithAddress := getBalanceCmd.String("address", "", "获取账户余额 ")
	//startNode 命令加参数
	flagMiner := startNodeCmd.String("miner", "", "定义挖矿奖励的地址。")

	switch os.Args[1] {
	case "createWallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createBlockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "addressList":
		err := addressListCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printChain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getBalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "resetUTXOSet":
		err := resetUTXOSetCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startNode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1)
	}
	//exit（0）：正常运行程序并退出程序；
	//exit（1）：非正常运行导致退出程序；
	// 如果命令被正常解析，执行以下操作
	// createWallet 命令解析
	if createWalletCmd.Parsed() {
		// 创建钱包
		cli.createWallet(nodeID)
	}

	// createBlockchain 命令解析
	if createBlockchainCmd.Parsed() {
		if IsValidForAddress([]byte(*flagCreateBlockchainWithAddress)) == false {
			fmt.Println("地址无效。")
			printUsage()
			os.Exit(1)
		}
		cli.createGenesisBlockchain(*flagCreateBlockchainWithAddress, nodeID)
	}

	// 输出addresslist
	if addressListCmd.Parsed() {
		// 创建钱包
		cli.addressList(nodeID)
	}
	// send 命令解析
	if sendBlockCmd.Parsed() {
		if *flagFrom == "" || *flagTo == "" || *flagAmount == "" {
			printUsage()
			os.Exit(1)
		}
		from := JSONToArray(*flagFrom)
		to := JSONToArray(*flagTo)
		// 检验地址是否合法
		for index, fromAddress := range from {
			if IsValidForAddress([]byte(fromAddress)) == false || IsValidForAddress([]byte(to[index])) == false {
				fmt.Println("地址无效。")
				printUsage()
				os.Exit(1)
			}
		}

		amount := JSONToArray(*flagAmount)
		cli.send(from, to, amount, nodeID, *flagMine)
	}

	// printChain 命令解析
	if printChainCmd.Parsed() {
		//fmt.Println("输出所有区块的信息")
		cli.printChain(nodeID)
	}

	// getBalance 命令解析
	if getBalanceCmd.Parsed() {
		if IsValidForAddress([]byte(*getBalanceWithAddress)) == false {
			fmt.Println("地址无效。")
			printUsage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceWithAddress, nodeID)
	}
	// resetUTXOSet 命令解析
	if resetUTXOSetCmd.Parsed() {
		cli.resetUTXOSet(nodeID)
		//cli.getBalance(*getBalanceWithAddress)
	}
	//  startNode命令解析
	if startNodeCmd.Parsed() {
		cli.startNode(nodeID, *flagMiner)
		//cli.getBalance(*getBalanceWithAddress)
	}
}

// 防止命令行报错
func isValidArgs() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1) // 退出，否则会报错
	}
}
