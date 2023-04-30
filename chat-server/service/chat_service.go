package service

import (
	"bytes"
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/repo"
	"chat-ai/chat-server/utils"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func getPrompt(data *model.RequestMessages) string {
	prompt := "The following is a conversation with an AI assistant named 小鱼儿. The assistant is helpful, creative, clever, and very friendly."
	var err error
	if data.Role != "" {
		prompt, err = repo.QueryPromptByRole(data.Role)
		if prompt == "" || err != nil {
			mylog.Logger.Errorf("查询 role: %s 失败:%v", data.Role, err)
			prompt = "The following is a conversation with an AI assistant named 小鱼儿. The assistant is helpful, creative, clever, and very friendly."
		}
	}
	return prompt
}

func transferMessage(data *model.RequestMessages) (*model.ChatRequest, int) {
	chatRequest := model.ChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: make([]model.ChatMessage, 0),
	}
	assistant := model.ChatMessage{
		Role: "system",
		//Content: "You are a chat assistant named 小鱼儿",
		Content: getPrompt(data),
	}
	chatRequest.Messages = append(chatRequest.Messages, assistant)
	cnt := 0
	for _, message := range data.Messages {
		if message.Ask != "" {
			msg1 := model.ChatMessage{
				Role:    "user",
				Content: message.Ask,
			}
			chatRequest.Messages = append(chatRequest.Messages, msg1)
		}

		if message.Answer != "" {
			cnt++
			msg2 := model.ChatMessage{
				Role:    "assistant",
				Content: message.Answer,
			}
			chatRequest.Messages = append(chatRequest.Messages, msg2)
		}
	}

	limit := config.GlobalConf.OpenAi.ChatLimit
	if cnt > limit {
		toClear := cnt - limit
		for i := 0; i < len(chatRequest.Messages) && toClear > 0; i++ {
			if chatRequest.Messages[i].Role == "assistant" {
				chatRequest.Messages[i].Content = ""
				toClear--
			}
		}
		mylog.Logger.Infof("已达到交流限制,%d，开始清理对话", limit)
	}
	return &chatRequest, cnt
}

func ChatWithContext(messages *model.RequestMessages) (*model.ChatResult, error) {
	startTime := time.Now()
	url := config.GlobalConf.OpenAi.Url
	var result = &model.ChatResult{}

	apiKey, err := utils.GetAppKey()
	if err != nil {
		result.Code = _const.REQUEST_FREQUENTLY
		return result, nil
	}

	chatReq, cnt := transferMessage(messages)
	//if cnt > config.GlobalConf.OpenAi.ChatLimit {
	//	mylog.Logger.Info("ask limit ", config.GlobalConf.OpenAi.ChatLimit)
	//	result.Article = "您已超过使用限制,请重新开启会话窗口。"
	//	return result, nil
	//}
	mylog.Logger.Infof("当前对话数: %d", cnt)

	text := chatReq.Messages[len(chatReq.Messages)-1].Content
	b, err := DetectText(text, apiKey)
	if err != nil {
		mylog.Logger.Errorf("检测文本失败：%s,  err:%v", text, err)
	}
	if b {
		mylog.Logger.Infof("检测到文本违禁：%s, err= %v", text, err)
		result.Article = _const.VIOLATE_CONTENT
		return result, nil
	}

	payload, err := json.Marshal(chatReq)
	if err != nil {
		mylog.Logger.Info("Error marshaling JSON:", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		mylog.Logger.Info("Error creating request:", err)
		return nil, err
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
				mylog.Logger.Errorf("Error sending request: %v status: %d", resErr, resp.StatusCode)
			} else {
				mylog.Logger.Error("Error sending request:", resErr)
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
		return nil, resErr
	}

	for _, choice := range chatResp.Choices {
		//s := strings.Replace(choice.Message.Content, "\n", "", -1)
		s := choice.Message.Content
		result.Article = s
	}
	result.ElapsedTime = time.Since(startTime).Milliseconds()

	return result, nil
}

func Chat(text string) *model.ChatResult {
	startTime := time.Now()
	url := config.GlobalConf.OpenAi.Url
	apiKey, err := utils.GetAppKey()
	if err != nil {
		return nil
	}
	var result = &model.ChatResult{}

	b, err := DetectText(text, apiKey)
	if err != nil {
		mylog.Logger.Errorf("检测文本失败：%s,  err:%v", text, err)
	}
	if b {
		mylog.Logger.Infof("检测到文本违禁：%s, err= %v", text, err)
		result.Article = _const.VIOLATE_CONTENT
		return result
	}

	chatReq := model.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []model.ChatMessage{
			{
				Role:    "system",
				Content: "You are a chat assistant named 小鱼儿",
			},
			{
				Role:    "user",
				Content: text,
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
			mylog.Logger.Errorf("Error reading response body: %v", resErr)
			continue
		}
		break
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if resErr != nil {
		mylog.Logger.Errorf("Error reading response body: %v", resErr)
		return result
	}
	//mylog.Logger.Infof("%+v", chatResp)
	for _, choice := range chatResp.Choices {
		//s := strings.Replace(choice.Message.Content, "\n", "", -1)
		result.Article = choice.Message.Content
	}
	result.ElapsedTime = time.Since(startTime).Milliseconds()

	return result
}

func copyBody(src io.ReadCloser) ([]byte, error) {
	b, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}
	src.Close()
	return b, nil
}

func resetBody(request *http.Request, originalBody []byte) {
	request.Body = io.NopCloser(bytes.NewBuffer(originalBody))
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(originalBody)), nil
	}
}

func QueryAllRoles(userId uint) ([]model.RolePromptVo, error) {
	prompt, err := repo.QueryAllRoleByUserId(userId)
	if err != nil {
		mylog.Logger.Errorf("获取prompt列表失败")
		return nil, err
	}
	var promptVo = make([]model.RolePromptVo, 0)
	for _, v := range prompt {
		var vo = model.RolePromptVo{
			Role:        v.Role,
			Description: v.Description,
		}
		promptVo = append(promptVo, vo)
	}

	return promptVo, nil
}
