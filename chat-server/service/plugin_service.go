package service

import (
	"bytes"
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/utils"
	"encoding/json"
	"net/http"
	"time"
)

func AskWithPrompt(prompt, question string) *model.PluginResult {
	url := config.GlobalConf.OpenAi.Url
	apiKey, err := utils.GetAppKey()
	if err != nil {
		return nil
	}
	var result = &model.PluginResult{}

	b, err := DetectText(question, apiKey)
	if err != nil {
		mylog.Logger.Errorf("检测文本失败：%s,  err:%v", question, err)
	}
	if b {
		mylog.Logger.Infof("检测到文本违禁：%s, err= %v", question, err)
		result.Answer = _const.VIOLATE_CONTENT
		return result
	}

	chatReq := model.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []model.ChatMessage{
			{
				Role:    "system",
				Content: prompt,
			},
			{
				Role:    "user",
				Content: question,
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
	var chatResp model.ChatResponse
	var originalBody []byte
	if req != nil && req.Body != nil {
		originalBody, _ = copyBody(req.Body)
		resetBody(req, originalBody)
	}
	var resp *http.Response
	var resErr error
	for i := 0; i < 3; i++ {
		resErr = nil
		resp, resErr = httpClient.Do(req)
		if resErr != nil || resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusTooManyRequests {
			if resp != nil {
				mylog.Logger.Errorf("Error sending code %v err:%v ", resp.StatusCode, resErr)
			} else {
				mylog.Logger.Errorf("Error sending resp %v err:%v ", resp, resErr)
			}
			// 重置body
			if req.Body != nil {
				resetBody(req, originalBody)
				resp.Body.Close()
			}
			time.Sleep(time.Second * 3)
			continue
		}
		// read the response body
		resErr = json.NewDecoder(resp.Body).Decode(&chatResp)
		if resErr != nil {
			mylog.Logger.Error("Error reading response body:", resErr)
			continue
		}
		break
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if resErr != nil {
		return result
	}

	for _, choice := range chatResp.Choices {
		//s := strings.Replace(choice.Message.Content, "\n", "", -1)
		result.Answer = choice.Message.Content
	}
	return result
}
