package service

import (
	"bytes"
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func GenerateArticle(text string) *model.ChatResult {
	startTime := time.Now()
	url := config.GlobalConf.OpenAi.Url
	apiKey, err := utils.GetAppKey()
	if err != nil {
		return nil
	}

	b, err := DetectText(text, apiKey)
	if err != nil {
		mylog.Logger.Errorf("检测文本失败：%s,  err:%v", text, err)
	}
	if b {
		mylog.Logger.Infof("检测到文本违禁：%s, err= %v", text, err)
		var result = &model.ChatResult{}
		result.Article = _const.VIOLATE_CONTENT
		return result
	}

	chatReq := model.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []model.ChatMessage{
			{
				Role:    "system",
				Content: "You are an article writing assistant named 小鱼儿",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Please write an article describing the following：\n %s", text),
			},
		},
	}
	payload, err := json.Marshal(chatReq)
	if err != nil {
		mylog.Logger.Error("Error marshaling JSON:", err)
		return nil
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		mylog.Logger.Error("Error creating request:", err)
		return nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := httpClient
	client.Timeout = 120 * time.Second
	var resp *http.Response
	var resErr error
	var chatResp model.ChatResponse

	for i := 0; i < 2; i++ {
		resp, resErr = client.Do(req)
		if resErr != nil || resp.StatusCode == http.StatusBadRequest {
			time.Sleep(time.Second)
			mylog.Logger.Error("Error sending request:", resErr)
			continue
		}
		// read the response body
		err = json.NewDecoder(resp.Body).Decode(&chatResp)
		if resErr != nil {
			mylog.Logger.Error("Error reading response body:", err)
			continue
		}
		break
	}
	defer resp.Body.Close()
	if resErr != nil {
		return nil
	}
	var result = &model.ChatResult{}
	for _, choice := range chatResp.Choices {
		//s := strings.Replace(choice.Message.Content, "\n", "", -1)
		result.Article = choice.Message.Content
	}
	result.ElapsedTime = time.Since(startTime).Milliseconds()

	return result
}
