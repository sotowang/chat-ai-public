package service

import (
	"bytes"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"encoding/json"
	"fmt"
	"net/http"
)

func DetectText(text string, apiKey string) (bool, error) {
	// 设置请求体
	input := map[string]string{"input": text}
	requestBody, err := json.Marshal(input)
	if err != nil {
		mylog.Logger.Error("json.Marshal error:", err)
		return false, err
	}

	// 发送POST请求

	url := "https://api.openai.com/v1/moderations"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		mylog.Logger.Error("http.NewRequest error:", err)
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		mylog.Logger.Error("client.Do error:", err)
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		mylog.Logger.Errorf("resp status code:%d ,apikei: %s", resp.StatusCode, apiKey)
		return false, fmt.Errorf("not ok error")
	}

	// 解析响应体
	var result model.ModerationResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		mylog.Logger.Error("json.NewDecoder error:", err)
		return false, err
	}

	return result.Results[0].Flagged, nil
}
