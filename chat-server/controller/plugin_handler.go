package controller

import (
	"bytes"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/repo"
	"chat-ai/chat-server/service"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func PluginAskHandler(c *gin.Context) {
	question, ok := c.GetPostForm("question")
	prompt, ok1 := c.GetPostForm("prompt")
	if !ok || question == "" || !ok1 || prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing text parameter"})
		return
	}
	mylog.Logger.Infof("user ask:[%s], prompt:[%s], question: [%s]",
		question, prompt, question)
	result := service.AskWithPrompt(prompt, question)
	if result != nil && result.Answer != "" {
		increaseQACount(_const.QA_TYPE_ASK, c)
	}
	c.JSON(200, result)
}

func increaseQACount(qaType int, c *gin.Context) {
	userID, _ := c.Get("userID")
	userId := userID.(uint)
	err := repo.IncreaseAskQACount(userId, qaType)
	if err != nil {
		mylog.Logger.Errorf("add user: %d ask count err: %v", userId, err)
	}
}

func PluginAppKeyHandler(c *gin.Context) {
	userID, _ := c.Get("userID")
	userId := userID.(uint)
	atoi := strconv.Itoa(int(userId))
	cyproto := AesEncrypt(atoi, _const.PLUGIN_SALT)
	mylog.Logger.Infof("userId: %d 获取appkey: %s ", userId, cyproto)
	c.JSON(200, gin.H{
		"appkey": cyproto,
	})
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
	return fmt.Sprintf("sk-%s", base64.StdEncoding.EncodeToString(cryted))
}

func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
