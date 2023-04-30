package controller

import (
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/repo"
	"chat-ai/chat-server/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

/**
检查pdf上传次数
*/
func checkPdfUploadLimit(c *gin.Context) error {
	userID, _ := c.Get("userID")
	userId := userID.(uint)
	//1.获取当前user今日使用pdf的次数
	count, err := repo.QueryTodayPdfRecord(userId)
	if err != nil {
		mylog.Logger.Errorf("查询用户: %d 今日pdf使用次数失败,err: %v", userId, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error})
		return err
	}
	if count >= config.GlobalConf.PDF.PdfCnt {
		mylog.Logger.Errorf("user id: %d pdf 使用次数为： %d， 超过：%d", userId, count, config.GlobalConf.PDF.PdfCnt)
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"err": "使用超限", "code": 42900})
		return fmt.Errorf("使用超限")
	}
	return nil
}

func ListDocController(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		return
	}
	userId := userID.(uint)
	docs, err := service.ListMyDocs(userId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{
			"docs":      docs,
			"doc_limit": config.GlobalConf.PDF.PdfCnt,
			"qa_limit":  config.GlobalConf.PDF.QACnt,
			"file_size": config.GlobalConf.PDF.Size,
			"formats":   config.GlobalConf.PDF.Formats,
			"about":     config.GlobalConf.PDF.About,
		})
}

func ProcessDocController(c *gin.Context) {
	err := checkPdfUploadLimit(c)
	if err != nil {
		return
	}
	userID, _ := c.Get("userID")
	userId := userID.(uint)
	var res = &model.UploadResponse{}
	//1.判断是否爬url
	url, ok := c.GetPostForm("url")
	if ok {
		mylog.Logger.Infof("要爬的url :%s", url)
		res, err = service.ProcessDoc(url, nil, url, "url", int(userId))
		if err != nil {
			if err == _const.INVALID_URL {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": "40000", "error:": "无效的url"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "40001", "error:": "爬取url失败"})
			return
		}
	} else {
		formFile, err := c.FormFile("file")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error:": "文件呢？", "code": "40400"})
			return
		}
		// 读取文件内容
		file, err := formFile.Open()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": "40400", "error:": "文件呢？"})
			return
		}
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": "40400", "error:": "读文件失败"})
			return
		}
		contentType := formFile.Header.Get("Content-Type")
		// 输出文件信息

		mylog.Logger.Printf("user id: %d 上传的文件名称: %s, 上传的文件大小: %d bytes",
			userId,
			formFile.Filename, len(fileBytes))
		if len(fileBytes) > config.GlobalConf.PDF.Size*1024*1024 {
			mylog.Logger.Printf("user id: %d 上传的文件名称: %s  太大", userId, formFile.Filename)
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"code": "41300", "error:": "上传的文件太大"})
			return
		}

		res, err = service.ProcessDoc("", fileBytes, formFile.Filename, contentType, int(userId))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "40400", "error:": "上传文件失败"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"res": res})

}

func UploadController(c *gin.Context) {
	formFile, err := c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error:": "文件呢？"})
		return
	}
	// 读取文件内容
	file, err := formFile.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": "40400", "error:": "文件呢？"})
		return
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": "40400", "error:": "读文件失败"})
		return
	}
	contentType := formFile.Header.Get("Content-Type")
	// 输出文件信息
	mylog.Logger.Printf("上传的文件名称: %s, 上传的文件大小: %d bytes", formFile.Filename, len(fileBytes))
	res, err := service.Upload(fileBytes, formFile.Filename, contentType)
	if err != nil {
		mylog.Logger.Errorf("上传文件失败: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "40400", "error:": "上传文件失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"res": res})
}

func DeleteController(c *gin.Context) {
	userID, ok := c.Get("userID")
	docId, ok1 := c.GetQuery("docId")
	if !ok || !ok1 {
		return
	}
	userId := userID.(uint)
	err := service.DeleteDoc(userId, docId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nil)
}

func checkPdfQACount(docId string, c *gin.Context) error {
	userID, _ := c.Get("userID")
	userId := userID.(uint)
	count, err := repo.QueryQACntByDocId(userId, docId)
	if err != nil {
		mylog.Logger.Errorf("查询用户:%d  docId: %s 使用 问答次数失败 err:%s", userId, docId, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return err
	}
	if count >= config.GlobalConf.PDF.QACnt {
		mylog.Logger.Errorf("用户:%d docId:%s 已经超出QA使用限制: %d", userId, docId, count)
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"err": "使用超限", "code": 42900})
		return fmt.Errorf("使用超限")
	}
	return nil
}

func AskController(c *gin.Context) {
	userID, _ := c.Get("userID")
	userId := userID.(uint)
	docId, ok1 := c.GetQuery("docId")
	question, ok2 := c.GetQuery("question")
	if !ok1 || !ok2 || docId == "" || question == "" {
		mylog.Logger.Errorf("docId: %s 或 question:%s 不存在", docId, question)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing  parameter"})
		return
	}
	mylog.Logger.Infof("user: %d pdf ask: %s", userId, question)
	err := checkPdfQACount(docId, c)
	if err != nil {
		return
	}

	res, err := service.Ask(int(userId), docId, question)
	if err != nil {
		mylog.Logger.Errorf("ask failed: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "40401", "error:": "网络错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"res": res})
}
