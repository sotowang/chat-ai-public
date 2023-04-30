package controller

import (
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	pay_server "chat-ai/chat-server/pay-server"
	"chat-ai/chat-server/repo"
	"chat-ai/chat-server/service"
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func TranslateHandler(c *gin.Context) {
	text, ok := c.GetPostForm("text")
	if !ok || text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing text parameter"})
		return
	}
	lan := c.PostFormArray("lan")
	mylog.Logger.Infof("翻译: text [%s], lan:[%s]", text, lan)

	result := service.Translatebychatgpt35(text, lan)

	if result != nil && result.TranslatedText != "" {
		increaseQACount(_const.QA_TYPE_ASK, c)
	}

	c.JSON(200, result)
}

func ArticleHandler(c *gin.Context) {
	text, ok := c.GetPostForm("text")
	if !ok || text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing text parameter"})
		return
	}
	mylog.Logger.Infof("生成文章: [%s]", text)
	result := service.GenerateArticle(text)
	c.JSON(200, result)
}

func ImageHandler(c *gin.Context) {
	text, ok := c.GetPostForm("text")
	if !ok || text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing text parameter"})
		return
	}
	result, _ := service.GenerateImage(text)
	mylog.Logger.Infof("image text:[%s]", text)
	if result != nil && result.Data != nil && len(result.Data) > 0 && result.Data[0].URL != "" {
		increaseQACount(_const.QA_TYPE_IMAGE, c)
	}
	c.JSON(200, result)
}

func ImageHandler2(c *gin.Context) {
	text, ok1 := c.GetPostForm("text")
	style, ok2 := c.GetPostForm("style")
	if !ok1 || !ok2 || text == "" || style == "" || len(text) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong text parameter"})
		return
	}
	mylog.Logger.Infof("image text:[%s],style: [%s]", text, style)

	result, _ := service.GenerateImage2(text, style)
	if result != nil && result.Data != nil && len(result.Data) > 0 && result.Data[0].URL != "" {
		increaseQACount(_const.QA_TYPE_IMAGE, c)
	}
	c.JSON(200, result)
}

func ChatHandler(c *gin.Context) {
	text, ok := c.GetPostForm("text")
	if !ok || text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing text parameter"})
		return
	}

	mylog.Logger.Infof("user ask:[%s]", text)
	result := service.Chat(text)
	if result != nil && result.Article != "" {
		increaseQACount(_const.QA_TYPE_ASK, c)
	}
	c.JSON(200, result)
}

func ChatCosplayHandler(c *gin.Context) {
	var data = &model.RequestMessages{}
	err := c.ShouldBindJSON(data)
	if err != nil || data.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing  parameter"})
		return
	}
	mylog.Logger.Infof("chat with cosplay:[%+v]", data)
	result, err := service.ChatWithContext(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.Code == _const.REQUEST_FREQUENTLY {
		c.JSON(_const.REQUEST_FREQUENTLY, gin.H{"error": "no useful appkeys"})
		return
	}
	if result != nil && result.Article != "" {
		increaseQACount(_const.QA_TYPE_ASK, c)
	}

	c.JSON(200, result)
}

func ChatWithContextHandler(c *gin.Context) {
	var data = &model.RequestMessages{}
	err := c.ShouldBindJSON(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing  parameter"})
		return
	}
	mylog.Logger.Infof("chat with context:[%+v]", data)
	result, err := service.ChatWithContext(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.Code == _const.REQUEST_FREQUENTLY {
		c.JSON(_const.REQUEST_FREQUENTLY, gin.H{"error": "no useful appkeys"})
		return
	}

	if result != nil && result.Article != "" {
		increaseQACount(_const.QA_TYPE_ASK, c)
	}
	c.JSON(200, result)
}

func SmsHandler(c *gin.Context) {
	email, ok1 := c.GetPostForm("email")
	isEmail := validateEmailFormat(email)
	isPhone := IsValidPhone(email)

	if !ok1 || email == "" || (!isEmail && !isPhone) {
		mylog.Logger.Errorf("email: %s 不存在", email)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing text parameter"})
		return
	}
	//如果用户使用的是手机号，则发短信
	if isPhone {
		err := service.SendMsg(email, config.GlobalConf.SMS.Sign, config.GlobalConf.SMS.TemplateCode)
		if err != nil {
			if err == _const.VERIFY_CODE_LIMIT {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "验证码超限，一天仅可获取3次"})
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "send msg wrong"})
			}
			return
		}
	} else {
		remain, err := service.SendEmailMsg(email)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "send msg wrong"})
			return
		}
		if remain > 0 {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"remain": remain,
				"code":   42900,
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func LoginByPhoneOrEmailHandler(c *gin.Context) {
	email, ok1 := c.GetPostForm("email")
	code, ok2 := c.GetPostForm("code")
	if !ok1 || !ok2 || email == "" || code == "" || (!IsValidPhone(email) && !validateEmailFormat(email)) {
		mylog.Logger.Errorf("email: %s 或code:%s 不存在", email, code)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing text parameter"})
		return
	}
	email = strings.ToLower(email)
	res, err := service.LoginByPhoneOrEmail(email, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "user": res})
}

func LoginHandler(c *gin.Context) {
	email, ok1 := c.GetPostForm("email")
	password, ok2 := c.GetPostForm("password")
	if !ok1 || !ok2 || email == "" || password == "" {
		mylog.Logger.Errorf("email: %s 或 password: %s 不存在", email, password)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing text parameter"})
		return
	}
	b, err := service.Login(email, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "user": b})
}

func IsValidPhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

func validateEmailFormat(email string) bool {
	// 正则表达式用于匹配常见的邮箱格式
	pattern := `^[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)*@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
	matched, _ := regexp.MatchString(pattern, email)
	invalidEmail := config.GlobalConf.SMS.BlackoutEmail

	for _, invalidSuffix := range invalidEmail {
		if strings.HasSuffix(email, invalidSuffix) {
			mylog.Logger.Infof("suffix: %s 无效的邮箱: %s", invalidSuffix, email)
			matched = false
			break
		}
	}
	return matched
}

func AboutHandler(c *gin.Context) {
	conf, err := config.LoadConfig()
	if err == nil {
		config.GlobalConf = conf
	}
	c.JSON(http.StatusOK, gin.H{
		"about": config.GlobalConf.About,
	})
}

func CosplayListHandler(c *gin.Context) {
	username, ok := c.GetQuery("username")
	userId := uint(0)
	if ok && username != "" {
		var existingUser = &model.User{}
		result := repo.QueryUserByPhoneOrEmail(username, existingUser)
		if result.Error == nil {
			userId = existingUser.ID
		}
	}

	mylog.Logger.Infof("userId: %d,CosplayListHandler", userId)
	roles, err := service.QueryAllRoles(userId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "网络错误",
		})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func RegisterHandler(c *gin.Context) {
	email, ok1 := c.GetPostForm("email")
	password, ok2 := c.GetPostForm("password")
	if !ok1 || !ok2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing text parameter"})
		return
	}
	b, err := service.Register(email, password, config.GlobalConf.Server.InitExpireTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, b)
}

func CallbackHandler(c *gin.Context) {
	//1.验证签名
	c.Request.ParseForm()
	ok, err := pay_server.VerifySign(c.Request.Form)
	if err != nil {
		mylog.Logger.Errorf("验证签名失败：%v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "验证签名失败"})
		return
	}
	if !ok {
		mylog.Logger.Errorf("签名验证不通过")
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "签名验证不通过"})
		return
	}
	//2.获取交易状态
	noti := pay_server.GetNotify(c.Request)
	if noti == nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	mylog.Logger.Infof("交易状态为：%s", noti.TradeStatus)

	if noti.TradeStatus == alipay.TradeStatusSuccess {
		_, err = service.UpdateOrderStatusAndAddUserTime(noti.OutTradeNo, _const.ORDER_SUCCESS_STATUS)
		if err != nil {
			mylog.Logger.Errorf("更新order: %s failed", noti.OutTradeNo)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "更新订单状态失败"})
			return
		}
	} else if noti.TradeStatus == alipay.TradeStatusClosed {
		err = service.UpdateOrderStatus(noti.OutTradeNo, _const.ORDER_FAILED_STATUS)
		if err != nil {
			mylog.Logger.Errorf("更新订单状态失败：%v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "更新订单状态失败"})
			return
		}
		mylog.Logger.Infof("更新订单：%s 状态为超时：%s", noti.OutTradeNo, noti.OutTradeNo)
	}
	// 3. 确认收到通知消息
	c.Writer.Write([]byte("success"))
}

func OrderHandler(c *gin.Context) {
	productIdStr, ok := c.GetPostForm("productId")
	userID, ok1 := c.Get("userID")
	source, ok2 := c.GetPostForm("source")
	payTypeStr, ok3 := c.GetPostForm("payType")
	if !ok1 || !ok || !ok2 || !ok3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameter"})
		return
	}
	userId := userID.(uint)
	productId, _ := strconv.Atoi(productIdStr)
	payType, _ := strconv.Atoi(payTypeStr)
	res, err := service.CreateOrder(uint64(userId), uint64(productId), 1, source, payType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func ProductHandler(c *gin.Context) {
	res, err := service.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)

}

func UserInfoHandler(c *gin.Context) {
	userID, ok1 := c.Get("userID")
	if !ok1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameter"})
		return
	}
	userId := userID.(uint)
	info, err := service.GetUserInfo(int(userId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "user": info})
}

func OrderInfoHandler(c *gin.Context) {
	userID, ok1 := c.Get("userID")
	if !ok1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing parameter"})
		return
	}
	userId := userID.(uint)
	orders, err := service.GetOrderByUserId(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "order": orders})

}
