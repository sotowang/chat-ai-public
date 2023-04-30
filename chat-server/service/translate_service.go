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
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient *http.Client

func InitClient() {
	httpClient = &http.Client{
		Transport: createProxyTransport(),
		Timeout:   time.Second * 45, // 超时设置
	}
}

func generatePrompt(text string, lan []string) string {
	languages := ""
	for i := 0; i < len(lan); i++ {
		languages += fmt.Sprintf("%d. %s ", i+1, lan[i])
	}
	if languages == "" {
		languages = "Chinese"
	}
	prompt := fmt.Sprintf("Please translate the sentences  into %s:\n\n '%s' ", languages, text)
	return prompt
}

func createProxyTransport() *http.Transport {
	proxyURL := config.GlobalConf.Server.ProxyUrl
	proxy := config.GlobalConf.Server.Proxy
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if proxy {
		proxyUrl, _ := url.Parse(proxyURL)
		transport.Proxy = http.ProxyURL(proxyUrl)
	}
	return transport
}

func Translatebychatgpt35(text string, lan []string) *model.TranslationResult {
	startTime := time.Now()
	url := config.GlobalConf.OpenAi.Url

	apiKey, err := utils.GetAppKey()
	if err != nil {
		return nil
	}

	prompt := generatePrompt(text, lan)
	b, err := DetectText(text, apiKey)
	if err != nil {
		mylog.Logger.Errorf("检测文本失败：%s,  err:%v", prompt, err)
	}
	if b {
		mylog.Logger.Infof("检测到文本违禁：%s, err= %v", prompt, err)
		var result = &model.TranslationResult{}
		result.TranslatedText = _const.VIOLATE_CONTENT
		return result
	}

	chatReq := model.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []model.ChatMessage{
			{
				Role:    "system",
				Content: "I would like you to translate the following sentence as a professional translator. Please only provide the translated sentence and do not say anything else.",
			},
			{
				Role:    "user",
				Content: prompt,
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
	var resp *http.Response
	var originalBody []byte
	if req != nil && req.Body != nil {
		originalBody, _ = copyBody(req.Body)
		resetBody(req, originalBody)
	}
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
			mylog.Logger.Errorf("Error reading response body: %v ", resErr)
			continue
		}
		break
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if resErr != nil {
		mylog.Logger.Errorf("Error reading response body: %v", resErr)
		return nil
	}
	var result = &model.TranslationResult{}
	for _, choice := range chatResp.Choices {
		s := strings.Replace(choice.Message.Content, "\n", "", -1)
		result.TranslatedText = s
	}
	result.ElapsedTime = time.Since(startTime).Milliseconds()

	return result
}

func Translate(text string, lan []string) {
	url := "https://api.openai.com/v1/completions"
	apiKey, err := utils.GetAppKey()
	if err != nil {
		return
	}
	prompt := generatePrompt(text, lan)
	data := model.CompletionRequest{
		Model:            "text-davinci-003",
		Prompt:           prompt,
		Temperature:      0.3,
		MaxTokens:        100,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		mylog.Logger.Error("Error marshaling JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		mylog.Logger.Error("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		mylog.Logger.Error("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// read the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		mylog.Logger.Error("Error reading response body:", err)
		return
	}

	mylog.Logger.Error(responseBody)
}
