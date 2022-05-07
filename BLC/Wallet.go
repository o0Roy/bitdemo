package BLC

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	ripemd1602 "golang.org/x/crypto/ripemd160"
	"log"
)

const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
	// 1. 私钥
	Privatekey ecdsa.PrivateKey

	// 2. 公钥
	Publickey []byte
}

// 获取钱包地址
func (w *Wallet) GetAddress() []byte {
	// 1. ripemd160
	// 20 字节
	ripemd150_sha256 := Ripemd160Sha256(w.Publickey)
	// 21 字节
	version_ripemd160_sha256 := append([]byte{version}, ripemd150_sha256...)
	// 2 次 256 hash
	checkSumBytes := CheckSum(version_ripemd160_sha256)
	// 25 字节
	bytes := append(version_ripemd160_sha256, checkSumBytes...)
	return Base58Encode(bytes)
}

func IsValidForAddress(address []byte) bool {
	// 说明：1、解密后截取最后 4 个字节
	//      2、前 21 个字节再进行双 Sha256 取最后 4 个字节
	//      3、两次的字节相同则有效
	// 25 字节
	version_public_checksumBytes := Base58Decode(address)
	slice_index := len(version_public_checksumBytes) - addressChecksumLen
	// 截取后面 4 个字节
	checkSumBytes := version_public_checksumBytes[slice_index:]
	// 截取前面 21 个字节
	version_ripemd160 := version_public_checksumBytes[:slice_index]
	// 2 次 256 hash
	checkBytes := CheckSum(version_ripemd160)
	if bytes.Compare(checkSumBytes, checkBytes) == 0 {
		return true
	}
	return false
}

// 返回公钥两次 sha256 后结果的后 4 个字节
func CheckSum(payload []byte) []byte {
	first_sha256 := sha256.Sum256(payload)
	second_sha256 := sha256.Sum256(first_sha256[:])
	return second_sha256[:addressChecksumLen]
}

func Ripemd160Sha256(publicKey []byte) []byte {
	// 1. sha256
	hash256 := sha256.New()
	hash256.Write(publicKey)
	hash := hash256.Sum(nil)

	// 2. ripemd160
	ripemd160 := ripemd1602.New()
	ripemd160.Write(hash)
	return ripemd160.Sum(nil)

}

// 产生新钱包
func NewWallet() *Wallet {
	privateKey, publicKey := newKeyPair()
	//fmt.Println(privateKey, publicKey)
	return &Wallet{privateKey, publicKey}
}

// 通过私钥，产生公钥
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	// 生成私钥
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	// 通过私钥生成公钥
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}
