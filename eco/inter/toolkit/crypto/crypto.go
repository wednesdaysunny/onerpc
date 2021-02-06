package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"sort"

	"golang.org/x/crypto/bcrypt"
)

const (
	numberchars = "1234567890"
)

func GetMD5Content(content string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(content))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func Base64Encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func Base64Decode(src string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(src)
}

func Base64URLEncode(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

func Base64URLDecode(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}

func RandInt63(max int64) int64 {
	var maxbi big.Int
	maxbi.SetInt64(max)
	value, _ := rand.Int(rand.Reader, &maxbi)
	return value.Int64()
}

func RandNumStr(l int) string {
	ret := make([]byte, 0, l)
	for i := 0; i < l; i++ {
		index := RandInt63(int64(len(numberchars)))
		ret = append(ret, numberchars[index])
	}
	return string(ret)
}

func RandBytes(c int) ([]byte, error) {
	b := make([]byte, c)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	} else {
		return b, nil
	}
}

func SimpleGuid() string {
	b := make([]byte, 24)
	binary.LittleEndian.PutUint64(b, uint64(time.Now().UnixNano()))
	if _, err := rand.Read(b[8:]); err != nil {
		return ""
	} else {
		return base64.RawURLEncoding.EncodeToString(b)
	}
}

func SaltPassword(password string) []byte {
	h, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		fmt.Println("Bcrypt salt password error: ", err)
		return nil
	}
	return h
}

func VerifyPassword(stored []byte, password string) bool {
	return bcrypt.CompareHashAndPassword(stored, []byte(password)) == nil
}

func SaltPassword2(password string) []byte {
	if rb, err := RandBytes(64); err != nil {
		return nil
	} else {
		pw := append(rb, []byte(password)...)
		hash := sha512.Sum512(pw)
		return append(rb, hash[:]...)
	}
}

func VerifyPassword2(stored []byte, password string) bool {
	if len(stored) != 128 {
		return false
	}
	pw := append([]byte{}, stored[:64]...)
	pw = append(pw, []byte(password)...)
	hash := sha512.Sum512(pw)
	return bytes.Equal(hash[:], stored[64:])
}

func Md5(data []byte) []byte {
	sum := md5.Sum(data)
	return sum[:]
}

func Md5Str(data []byte) string {
	return base64.StdEncoding.EncodeToString(Md5(data))
}

func Encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext := padding(data)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	return appendHmac(ciphertext, key), nil
}

func Decrypt(data, key []byte) ([]byte, error) {
	data, err := removeHmac(data, key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	// CBC mode always works in whole blocks.
	if len(data)%aes.BlockSize != 0 {
		errors.New("data is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(data))
	mode.CryptBlocks(plaintext, data)
	return unPadding(plaintext), nil
}

func padding(data []byte) []byte {
	blockSize := aes.BlockSize
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func unPadding(data []byte) []byte {
	l := len(data)
	unpadding := int(data[l-1])
	return data[:(l - unpadding)]
}

func appendHmac(data, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	hash := mac.Sum(nil)
	return append(data, hash...)
}

func removeHmac(data, key []byte) ([]byte, error) {
	if len(data) < sha256.Size {
		return nil, errors.New("Invalid length")
	}
	p := len(data) - sha256.Size
	mmac := data[p:]
	mac := hmac.New(sha256.New, key)
	mac.Write(data[:p])
	exp := mac.Sum(nil)
	if hmac.Equal(mmac, exp) {
		return data[:p], nil
	} else {
		return nil, errors.New("MAC doesn't match")
	}
}

func GenSignContent(dataMap map[string]string, secret string) string {
	listData := make([]string, 0, 10)
	for key, value := range dataMap {
		listData = append(listData, key+"="+value)
	}
	sort.Strings(listData)
	buffString := bytes.Buffer{}
	for _, sortString := range listData {
		buffString.WriteString(GetMD5Content(sortString))
	}
	buffString.WriteString(secret)
	return GetMD5Content(string(buffString.Bytes()))
}

// length(key) = 16 len(aes_iv) = 16
func AESEncrypterContent(content string, keyText, aesIv []byte) (result string) {
	result = ""
	if content == "" {
		return
	}
	block, err := aes.NewCipher(keyText) //选择加密算法
	if err != nil {
		return
	}
	contentBytes := []byte(content)
	contentBytes = PKCS7Padding(contentBytes, block.BlockSize())

	blockModel := cipher.NewCBCEncrypter(block, aesIv)
	ciphertext := make([]byte, len(contentBytes))

	blockModel.CryptBlocks(ciphertext, contentBytes)
	result = Base64Encode(ciphertext)
	return
}

// length(key) = 16 len(aes_iv) = 16
func AESDecrypterContent(ciphertext string, keyText, aesIV []byte) (result string) {
	result = ""
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			fmt.Println("Error:AESDecrypterContent", ciphertext)
		}
	}()
	if ciphertext == "" {
		return
	}
	ciphertestBytes, err := Base64Decode(ciphertext)
	if err != nil || len(ciphertestBytes) == 0 {
		return
	}
	block, err := aes.NewCipher(keyText) //选择加密算法
	if err != nil || block.BlockSize() <= 0 {
		return
	}
	blockModel := cipher.NewCBCDecrypter(block, aesIV)
	plantText := make([]byte, len(ciphertestBytes))
	blockModel.CryptBlocks(plantText, ciphertestBytes)
	plantText = PKCS7UnPadding(plantText, block.BlockSize())
	result = string(plantText)
	return
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(plantText []byte, blockSize int) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	step := length - unpadding
	if step < 0 || step > length {
		return []byte("")
	}
	return plantText[:(length - unpadding)]
}

func AesEncrypt(orig string, key string) string {
	// 转成字节数组
	origData := []byte(orig)
	k := []byte(key)
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)
	return base64.StdEncoding.EncodeToString(cryted)
}
func AesDecrypt(cryted string, key string) string {
	// 转成字节数组
	crytedByte, _ := base64.StdEncoding.DecodeString(cryted)
	k := []byte(key)
	// 分组秘钥
	block, _ := aes.NewCipher(k)
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig = PKCS7UnPadding(orig, blockSize)
	return string(orig)
}
