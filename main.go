package main

import (
	admin_server "chat-ai/admin-server"
	chat_server "chat-ai/chat-server"
	"chat-ai/chat-server/config"
	mylog "chat-ai/chat-server/log"
	pay_server "chat-ai/chat-server/pay-server"
	"chat-ai/chat-server/repo"
	"chat-ai/chat-server/service"
	"chat-ai/chat-server/utils"
	"flag"
	"fmt"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func main() {
	configFile := flag.String("config", "config-dev.ini", "path to config file")
	flag.Parse()
	var err error
	config.LoadCfg(*configFile)
	config.GlobalConf, err = config.LoadConfig()
	if err != nil {
		fmt.Println("无法读取配置文件: ", err)
		return
	}
	mylog.InitLog()

	//支付模块
	err = pay_server.CreatePayClient()
	if err != nil {
		mylog.Logger.Errorf("创建支付client失败:%v", err)
		panic(err)
	}

	//mysql
	repo.GetDB()
	if err != nil {
		panic(err)
	}

	//锁
	utils.AppKeyLock = utils.NewLock(3 * time.Second)

	//httpclient
	service.InitClient()

	//sms client
	err = service.InitSmsClient(config.GlobalConf.SMS.Ak, config.GlobalConf.SMS.Sk)
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(requestid.New(), AddLogInfo(), CORS, LogMiddleware)
	r.RedirectTrailingSlash = false
	r.GET("/", func(c *gin.Context) {
		c.String(200, "shi-yu")
	})

	chat_server.InitRouter(r)
	admin_server.InitRouter(r)

	port := fmt.Sprintf(":%d", config.GlobalConf.Server.Port)
	// 启动热更新协程
	go func() {
		hotUpdate()
	}()
	r.Run(port)

}

func hotUpdate() {
	for {
		time.Sleep(time.Duration(3) * time.Second)
		newConf, err := config.LoadConfig()
		if err != nil {
			mylog.Logger.Errorf("无法读取配置文件:%v ", err)
			continue
		}
		config.GlobalConf = newConf
	}
}

func LogMiddleware(c *gin.Context) {
	t := time.Now()
	c.Next()
	method := c.Request.Method
	reqUrl := c.Request.RequestURI
	statusCode := c.Writer.Status()
	clientIp := c.ClientIP()
	t2 := time.Since(t).Milliseconds()
	mylog.Logger.Infof("|%s|%s|%s|%d|%d ms", clientIp, method, reqUrl, statusCode, t2)
}

func AddLogInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqId := ctx.Request.Header.Get("X-Request-Id")
		device := ctx.Request.Header.Get("User-Agent")
		field := logrus.WithField("r", reqId).WithField("d", device)
		mylog.Logger = field
	}
}

func CORS(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	if c.Request.Method == "OPTIONS" {
		c.Abort()
		c.AbortWithStatus(http.StatusOK)
		return
	}
	c.Next()
}
