package BLC

import (
	"bytes"
	"math/big"
)

// base64

//ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/
//0(零)，O(大写的 o)，I(大写的i)，l(小写的 L)，+，/

// 加密种子
var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// 字节数组转 Base58 ，加密
// 1、加密的字符串越长，则加密后的信息越长。
// 2、该过程可以理解为转成58进制。
func Base58Encode(input []byte) []byte {
	//fmt.Println(input)
	var result []byte
	//var str string
	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}
	//var a int = 0
	for x.Cmp(zero) != 0 {
		// 该过程可以理解为转为58进制
		//a++
		//fmt.Println(&x)
		//fmt.Println(base)
		x.DivMod(x, base, mod)
		//z, _ :=x.DivMod(&x, base, mod)
		//fmt.Println(z)
		//fmt.Println(mod)
		//fmt.Println(mod.Int64())
		//fmt.Printf("%c\n", b58Alphabet[mod.Int64()])
		//fmt.Println(">>>>>>>>>>>")
		//s := fmt.Sprintf("%c", b58Alphabet[mod.Int64()])
		//str = str + s
		result = append(result, b58Alphabet[mod.Int64()])
	}
	//fmt.Println(a)
	//因为之前先附加低位的，后附加高位的，所以需要翻转
	ReverseBytes(result)
	//fmt.Println("start")
	//fmt.Println(input)
	for b := range input {
		if b == 0x00 {
			result = append([]byte{b58Alphabet[0]}, result...)
		} else {
			break
		}
	}
	//fmt.Println(result)
	return result
}

// Base58 转字节数组，解密
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)
	zeroBytes := 0
	for b := range input {
		if b == 0x00 {
			zeroBytes++
		}
	}
	payload := input[zeroBytes:]
	for _, b := range payload {
		// 类似58进制，转10进制。
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}
	decoded := result.Bytes()
	decoded = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), decoded...)
	return decoded
}
