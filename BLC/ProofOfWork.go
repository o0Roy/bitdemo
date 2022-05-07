package BLC

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

// sha256，一共256位

const TARGET_BIT = 16 // 假设256位 hash 前面至少有16个零

type ProofOfWork struct {
	Block  *Block   // 当前要验证的区块
	target *big.Int // 大数据存储
}

// 创建新的工作量证明对象
func NewProofOfWork(block *Block) *ProofOfWork {
	// 1.big.Int 对象 1
	// Poof_of_work 思路
	// 假设 hash 总共八位，难度为2（即 hash 值前两位为0）
	// 0000 0000
	// 0100 0000  target_val = 2^6 = 64
	// 0010 0000  find_hash_val 小于等于32
	// 只要 find_hash_val <= target_val  ==> 满足前两个0的要求
	// 同理可推广成256位的 hash 值
	// 0000 0000 0000 0000 0000 ..... 0000 0000
	// 1. 创建一个初始值为1的 target
	target := big.NewInt(1)
	// 2. 左移 256 - target_bit
	// Lsh sets z = x << n and returns z. 左移一位 等于 * 2
	// eg：难度为2 长度为8 即target = 0100 0000
	target.Lsh(target, 256-TARGET_BIT)
	//s := fmt.Sprintf("%0256b", target)
	//fmt.Println(s)
	//fmt.Printf("左移%d位\n", 256-TARGET_BIT)
	//fmt.Println(len(target.String()))
	return &ProofOfWork{block, target}
}

func (proofOfWork *ProofOfWork) Run() ([]byte, int64) {
	// 1. 将 Block 的属性拼接成字节数组
	// 2. 生成 hash
	// 3. 判断 hash 有效性
	nonce := 0
	var hashInt big.Int // 存储新生成的 hash
	var hash [32]byte
	for {
		// 准备数据
		dataBytes := proofOfWork.prepareData(nonce)
		// 生成 hash
		hash = sha256.Sum256(dataBytes)
		//fmt.Printf("\r%x\n", hash)
		// 将 hash 存储到 hashInt
		//fmt.Println(len(hash))
		// 每个字节(即每个字节ASCII码）用两字符十六进制数表示 原长度 32 输出后 64 x 小写输出 X 大写输出
		//fmt.Printf("%x\n", string(hash[:]))
		// 设置为无符号字节
		hashInt.SetBytes(hash[:]) // hash 转 big.Int 比较
		//fmt.Println(hashInt)
		// fmt.Println(hashInt)
		// 判断 hashInt 是否小于 Block 里面的 target
		// str := "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001766847064778384329583297500742918515827483896875618958121606201292619776"

		//fmt.Printf("%0256d\n", proofOfWork.target)
		//fmt.Println(len(str))

		if proofOfWork.target.Cmp(&hashInt) == 1 { // 小于 -1 等于 0 大于 1
			//str := fmt.Sprintf("%0256d", proofOfWork.target)
			//str1 := fmt.Sprintf("%0256d", )
			//fmt.Println(str)
			//fmt.Println(strings.Index(str, "1"))

			//转成二进制查看
			//s := fmt.Sprintf("%0256b", &hashInt)
			//fmt.Println(s)
			break
		}
		nonce = nonce + 1
	}
	// fmt.Printf("\r%x\n", hash)
	//fmt.Println(nonce)
	return hash[:], int64(nonce)
}

// 数据拼接，返回字节数组
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,           // 上一个区块 hash
			pow.Block.HashTransactions(),      // 本区块数据
			IntToHex(pow.Block.Timestamp),     // 时间戳
			IntToHex(int64(TARGET_BIT)),       // 目标位（难度）所有区块都加了，所以加不加无所谓
			IntToHex(int64(nonce)),            // 拼接位
			IntToHex(int64(pow.Block.Height)), // 区块高度
		},
		[]byte{},
	)
	return data
}
