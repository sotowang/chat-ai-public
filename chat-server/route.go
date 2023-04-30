package chat_server

import (
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	"chat-ai/chat-server/controller"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/repo"
	"chat-ai/chat-server/utils"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	time "time"
)

func InitRouter(r *gin.Engine, handlers ...gin.HandlerFunc) {
	r.Use(logRequest)
	auth := r.Group("/")

	// 不需要鉴权的路由
	auth.POST("/login/", controller.LoginByPhoneOrEmailHandler)
	auth.POST("/getVerifyCode/", controller.SmsHandler)
	auth.POST("/register/", controller.RegisterHandler)
	auth.POST("/call_back", controller.CallbackHandler)
	auth.GET("/about/", controller.AboutHandler)
	auth.GET("/cosplayList/", controller.CosplayListHandler)

	pluginRouter := auth.Group("/", handlers...)
	pluginRouter.POST("/plugin/ask/", pluginMiddleware(), checkQAExpire(_const.QA_TYPE_ASK), controller.PluginAskHandler)
	pluginRouter.POST("/plugin/chatContext/", pluginMiddleware(), checkQAExpire(_const.QA_TYPE_ASK), controller.ChatWithContextHandler)

	// 需要鉴权的路由
	router := auth.Group("/", handlers...)
	router.Use(authMiddleware())
	router.POST("/translate/", checkQAExpire(_const.QA_TYPE_ASK), controller.TranslateHandler)
	router.POST("/article/", checkQAExpire(_const.QA_TYPE_ASK), controller.ArticleHandler)
	router.POST("/image/", checkQAExpire(_const.QA_TYPE_IMAGE), controller.ImageHandler)
	router.POST("/image2/", checkQAExpire(_const.QA_TYPE_IMAGE), controller.ImageHandler2)
	router.POST("/chat/", checkQAExpire(_const.QA_TYPE_ASK), controller.ChatHandler)
	router.POST("/chatContext/", checkQAExpire(_const.QA_TYPE_ASK), controller.ChatWithContextHandler)
	router.POST("/chatCosplay/", checkQAExpire(_const.QA_TYPE_ASK), controller.ChatCosplayHandler)
	router.GET("/plugin/appKey/", controller.PluginAppKeyHandler)

	router.POST("/createOrder/", controller.OrderHandler)
	router.GET("/products/", controller.ProductHandler)

	router.GET("/userInfo/", controller.UserInfoHandler)
	router.GET("/orderInfo/", controller.OrderInfoHandler)

	router.POST("/processDoc/", checkExpire(), controller.ProcessDocController)
	router.GET("/ask/", checkExpire(), controller.AskController)
	router.DELETE("/delete/", controller.DeleteController)
	router.GET("/pdfList/", controller.ListDocController)

}

func checkExpire() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("userID")
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing parameter"})
			return
		}
		userId := userID.(uint)
		var user = &model.User{}
		result := repo.QueryById(int(userId), user)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error})
			return
		}
		if user.VipExpireDate <= time.Now().Unix() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "账户余额不足"})
			return
		}
		if user.Status == _const.USER_BANNED_STATUS {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户被封禁"})
			return
		}
	}
}

func checkQAExpire(qaType int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("userID")
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing parameter"})
			return
		}
		userId := userID.(uint)
		var user = &model.User{}
		result := repo.QueryById(int(userId), user)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error})
			return
		}
		if user.VipExpireDate <= time.Now().Unix() {
			//如果是非vip用户，则判断免费额度
			expire, _ := checkNoVipExpire(userId, qaType)
			if expire {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "账户余额不足"})
				return
			}
		}
		if user.Status == _const.USER_BANNED_STATUS {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户被封禁"})
			return
		}
	}
}

func checkNoVipExpire(userId uint, qaType int) (bool, error) {
	count, err := repo.QueryTodayQARecord(userId, qaType)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		mylog.Logger.Errorf("查询user：%d ask count err：%v", userId, err)
		return true, err
	}
	if qaType == _const.QA_TYPE_ASK && count >= config.GlobalConf.Chat.UserQALimitCnt {
		mylog.Logger.Infof("查询user：%d ask count ：%v ,limit: %d", userId, count, config.GlobalConf.Chat.UserQALimitCnt)
		return true, nil
	}

	if qaType == _const.QA_TYPE_IMAGE && count >= config.GlobalConf.Chat.UserImageLimitCnt {
		mylog.Logger.Infof("查询user：%d ask count ：%v ,limit: %d", userId, count, config.GlobalConf.Chat.UserImageLimitCnt)
		return true, nil
	}

	return false, nil
}

func pluginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		userId, err := AesDecrypt(tokenString, _const.PLUGIN_SALT)
		if err != nil {
			mylog.Logger.Errorf("解析 appkey: %s 失败:%v", tokenString, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		c.Set("userID", userId)
		mylog.Logger.Infof("当前操作的用户id: %d", userId)
		c.Next()
	}
}

func AesDecrypt(cryted string, key string) (uint, error) {
	prefix := "sk-"
	if !strings.HasPrefix(cryted, prefix) {
		return 0, errors.New("invalid appkey")
	}
	cryted = strings.Replace(cryted, prefix, "", 1)
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
	orig = PKCS7UnPadding(orig)
	userInt, err := strconv.Atoi(string(orig))
	if err != nil {
		return 0, err
	}

	return uint(userInt), nil
}

//去码
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		refreshToken := c.Request.Header.Get("RefreshToken")
		if authHeader == "" || refreshToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		refreshtTokenString := strings.Replace(refreshToken, "Bearer ", "", 1)
		claims, err := utils.ParseJWT(tokenString)

		if err != nil {
			mylog.Logger.Errorf("token error:%v ,token: %s", err, tokenString)
			if err.(*jwt.ValidationError).Errors == 16 {
				newAccessToken, err := utils.RefreshToken(c, tokenString, refreshtTokenString)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
					return
				}
				//c.Set("AccessToken", newAccessToken)
				c.Writer.Header().Set("Access-Token", newAccessToken)
				c.Next()
				return
			}
		}
		userID, ok := claims["userID"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing userID claim"})
			return
		}
		c.Set("userID", uint(userID))
		mylog.Logger.Infof("当前操作的用户id: %d", uint(userID))
		c.Next()
	}
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func logRequest(c *gin.Context) {
	requestID := generateRequestID()
	c.Set("requestID", requestID)
	c.Header("X-Request-ID", requestID)
	mylog.Logger.WithFields(logrus.Fields{
		//"requestID": requestID,
		"m":  c.Request.Method,
		"p":  c.Request.URL.Path,
		"ip": c.ClientIP(),
	}).Info("Received request")
}
