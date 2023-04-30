package service

import (
	"bytes"
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/repo"
	"chat-ai/chat-server/utils"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func ProcessDoc(scrawlUrl string, data []byte, originFilename string, contentType string, userId int) (*model.UploadResponse, error) {
	apiUrl := fmt.Sprintf("%s/process", config.GlobalConf.PDF.Server)
	req, err := http.NewRequest("POST", apiUrl, nil)
	if err != nil {
		mylog.Logger.Errorf("构建req失败: %v", err)
		return nil, err
	}
	var result = &model.UploadResponse{}
	appkey, err := utils.GetDocAppKey()
	if err != nil {
		return nil, err
	}
	// 创建一个multipart表单
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	outFilename := ""
	docId := ""
	if scrawlUrl != "" {
		//check apiUrl
		u, err := url.Parse(scrawlUrl)
		if !(err == nil && u.Scheme != "" && u.Host != "") {
			mylog.Logger.Errorf("无效的url: %s", scrawlUrl)
			return nil, _const.INVALID_URL
		}
		docId = getHash([]byte(scrawlUrl))
		outFilename = fmt.Sprintf("%s/%d-%s", config.GlobalConf.PDF.Dir, userId, docId)

		_ = writer.WriteField("input_file_path", scrawlUrl)
		_ = writer.WriteField("out_file_name", outFilename)
		_ = writer.WriteField("app_key", appkey)
		writer.Close()
	} else {
		//1.获取hash值
		docId = getHash(data)
		outFilename = fmt.Sprintf("%s/%d-%s", config.GlobalConf.PDF.Dir, userId, docId)
		ext := filepath.Ext(originFilename)
		filePath := fmt.Sprintf("%s%s", outFilename, ext)
		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// 文件不存在，写入文件
			if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
				mylog.Logger.Errorf("写入文件失败: %v", err)
				return nil, err
			}
		} else {
			// 文件存在，不需要写入
			mylog.Logger.Infof("文件 %s 已存在，不需要写入\n", filePath)
		}
		// 创建一个multipart表单
		_ = writer.WriteField("input_file_path", filePath)
		_ = writer.WriteField("out_file_name", outFilename)
		_ = writer.WriteField("app_key", appkey)
		writer.Close()
	}

	// 设置请求头和请求体
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Body = ioutil.NopCloser(body)

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		mylog.Logger.Errorf("request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应结果
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil || fmt.Sprintf("%s", data) == "fail" {
		mylog.Logger.Errorf("read error: %v", err)
		return nil, fmt.Errorf("process error")
	}
	result.DocId = docId
	result.Message = fmt.Sprintf("%s", data)
	err = recordUploadPdf(docId, contentType, originFilename, userId)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			mylog.Logger.Errorf("记录文件: %s 已存在", outFilename)
			return result, nil
		}
		mylog.Logger.Errorf("记录文件: %s 到数据库失败", outFilename)
		return nil, err
	}
	return result, nil
}

func deleteFilesByDocId(userId uint, docId string) error {
	dirPath := config.GlobalConf.PDF.Dir
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	fileNames, err := dir.Readdirnames(0)
	if err != nil {
		mylog.Logger.Errorf("读取文件目录 %s 失败", dirPath)
		return err
	}

	for _, fileName := range fileNames {
		if strings.HasPrefix(fileName, fmt.Sprintf("%d-%s", userId, docId)) {
			filePath := filepath.Join(dirPath, fileName)
			err := os.Remove(filePath)
			if err != nil {
				mylog.Logger.Errorf("删除文件 %s 失败", filePath)
				continue
			}
			mylog.Logger.Infof("删除文件 %s 成功", filePath)

		}
	}
	return nil
}

func DeleteDoc(userId uint, docId string) error {
	deleteFilesByDocId(userId, docId)
	err := repo.SetPdfRecordStatusToDeleted(userId, docId)
	if err != nil {
		mylog.Logger.Errorf("delete docId: %s err:%v", docId, err)
		return err
	}
	return nil
}

func ListMyDocs(userId uint) ([]model.DocsVO, error) {
	records, err := repo.QueryPdfRecordByStatus(userId, 0, 50)
	if err != nil {
		mylog.Logger.Errorf("get docs error: %v", err)
		return nil, err
	}
	var recordVos = make([]model.DocsVO, 0)
	for _, val := range records {
		record := model.DocsVO{
			DocID:    val.DocID,
			Filename: val.Filename,
		}
		recordVos = append(recordVos, record)
	}
	return recordVos, nil
}

func recordUploadPdf(docId, docType, filename string, userId int) error {
	err := repo.SaveUploadPdfRecord(userId, docId, filename, docType)
	return err
}

func getHash(data []byte) string {
	// 获取文件hash
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func Upload(data []byte, filename, contentType string) (*model.UploadResponse, error) {
	// 服务端接口地址
	url := "http://localhost:5174/api/upload"

	// 创建一个空的bytes.Buffer，用于构建multipart/form-data请求体
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	//
	//fileWriter, err := writer.CreateFormFile("file", filename)
	//if err != nil {
	//	mylog.Logger.Errorf("创建文件写入器失败：%v", err)
	//	return nil, err
	//}
	fileWriter, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":        []string{contentType},
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename)},
	})
	if err != nil {
		mylog.Logger.Errorf("创建文件写入器失败：%v", err)
		return nil, err
	}
	// 将要上传的文件写入file writer中
	_, err = io.Copy(fileWriter, bytes.NewReader(data))
	if err != nil {
		mylog.Logger.Errorf("文件写入失败：%v", err)
		return nil, err
	}

	// 必须手动调用multipart.Writer.Close()方法，以结束multipart/form-data请求体的写入
	writer.Close()

	// 创建POST请求
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		mylog.Logger.Errorf("创建请求失败：%v", err)
		return nil, err
	}
	// 设置请求头部中的Content-Type字段
	//req.Header.Set("Content-Type", writer.FormDataContentType())
	//req.Header.Set("Content-Type", contentType)

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		mylog.Logger.Errorf("发送请求失败：%v,status :%s", err, resp.Status)
		return nil, fmt.Errorf("upload fail")
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mylog.Logger.Errorf("读取响应失败：%v", err)
		return nil, err
	}
	var response = &model.UploadResponse{}
	if err = json.Unmarshal(respBody, response); err != nil {
		mylog.Logger.Errorf("解码响应失败：%+v", respBody)
		return nil, err
	}
	// 输出响应
	if response != nil && response.Message == "success" {
		mylog.Logger.Infof("上传文件: %s 成功", response.Message)
		return response, err
	}
	mylog.Logger.Errorf("上传文件: %+v 失败", response)
	return nil, fmt.Errorf("上传文件失败")
}

func Ask(userId int, docId, question string) (*model.AskResponse, error) {

	// 创建表单数据
	body := &bytes.Buffer{}
	outFilename := fmt.Sprintf("%s/%d-%s", config.GlobalConf.PDF.Dir, userId, docId)
	writer := multipart.NewWriter(body)
	appKey, err := utils.GetDocAppKey()
	if err != nil {
		return nil, err
	}
	_ = writer.WriteField("query", question)
	_ = writer.WriteField("out_file_name", outFilename)
	_ = writer.WriteField("app_key", appKey)
	_ = writer.Close()
	url := fmt.Sprintf("%s/ask", config.GlobalConf.PDF.Server)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Println("创建请求失败：", err)
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("发送请求失败：", err)
		return nil, err
	}

	// 解析响应
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		mylog.Logger.Errorf("ask failed, err:%v,data: %s", err, data)
		return nil, fmt.Errorf("internal error")
	}
	var res model.Response
	err = json.Unmarshal(data, &res)
	if err != nil {
		// 解析出错
		mylog.Logger.Errorf("unmarshar data:%s err:%v", data, err)
		return nil, fmt.Errorf("internal error")
	}
	var askRes = &model.AskResponse{
		Msg:   res.Answer,
		DocID: docId,
	}
	//记录提问数量
	err = repo.IncreasePdfQACount(userId, docId)
	if err != nil {
		mylog.Logger.Errorf("userId: %d, docId:%s 记录提问数量失败:err:%v", userId, docId, err)
	}
	return askRes, nil
}
