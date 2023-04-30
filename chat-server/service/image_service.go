package service

import (
	"bytes"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type TaskRequest struct {
	IsPremium bool              `json:"is_premium"`
	InputSpec map[string]string `json:"input_spec"`
}

type TaskResponse struct {
	ID string `json:"id"`
}

type TaskStateResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	State     string    `json:"state"`
	InputSpec InputSpec `json:"input_spec"`
	Premium   bool      `json:"premium"`
	//CreatedAt    time.Time `json:"created_at"`
	//UpdatedAt    time.Time `json:"updated_at"`
	IsNSFW       bool     `json:"is_nsfw"`
	PhotoURLList []string `json:"photo_url_list"`
	Result       struct {
		Final string `json:"final"`
	} `json:"result"`
}

type InputSpec struct {
	GenType           string `json:"gen_type"`
	Style             int    `json:"style"`
	Prompt            string `json:"prompt"`
	AspectRatioWidth  int    `json:"aspect_ratio_width"`
	AspectRatioHeight int    `json:"aspect_ratio_height"`
}

type TokenResp struct {
	FederatedID   string `json:"federatedId"`
	ProviderID    string `json:"providerId"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	PhotoURL      string `json:"photoUrl"`
	LocalID       string `json:"localId"`
	IDToken       string `json:"idToken"`
	Context       string `json:"context"`
	RefreshToken  string `json:"refreshToken"`
	ExpiresIn     string `json:"expiresIn"`
	OauthIDToken  string `json:"oauthIdToken"`
	RawUserInfo   string `json:"rawUserInfo"`
	Kind          string `json:"kind"`
}

func getToken() (string, error) {
	url := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=AIzaSyDCvp5MTJLUdtBYEKYWXJrlLzu1zuKM6Xw"
	payload := struct {
		Email             string `json:"email"`
		Password          string `json:"password"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		Email:             "xxx",
		Password:          "xxx",
		ReturnSecureToken: true,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		mylog.Logger.Errorf("get token error: %v", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		mylog.Logger.Errorf("get token error: %s", resp.Body)
		return "", fmt.Errorf("get token error")
	}
	var tokenResp TokenResp
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)

	if err != nil {
		// 解析错误处理
		return "", err
	}
	return tokenResp.IDToken, nil
}
func GenerateImage2(prompt, style string) (*model.GenerationResponse, error) {
	// 设置请求参数
	taskReq := TaskRequest{
		IsPremium: false,
		InputSpec: map[string]string{
			"prompt":       prompt,
			"style":        style,
			"display_freq": "10",
		},
	}

	// 将请求参数编码为 JSON 格式
	reqBody, err := json.Marshal(taskReq)
	if err != nil {
		fmt.Println("JSON encode error:", err)
		return nil, err
	}

	// 发送创建任务的请求
	req, err := http.NewRequest("POST", "https://paint.api.wombo.ai/api/v2/tasks", bytes.NewBuffer(reqBody))
	if err != nil {
		mylog.Logger.Error("NewRequest error:", err)
		return nil, err
	}
	token, err := getToken()
	if err != nil || token == "" {
		mylog.Logger.Error("get token error:", err)
		return nil, err
	}
	token = fmt.Sprintf("bearer %s", token)
	//mylog.Logger.Infof("获取token：%s", token)
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		mylog.Logger.Error("Request error:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应结果
	var taskResp TaskResponse
	err = json.NewDecoder(resp.Body).Decode(&taskResp)
	if err != nil || taskResp.ID == "" {
		mylog.Logger.Errorf("JSON decode error:%v or taskId nil", err)
		return nil, err
	}
	mylog.Logger.Info("Task ID:", taskResp.ID)

	result := &model.GenerationResponse{}
	result.Created = time.Now().Unix()
	result.Data = make([]struct {
		URL string `json:"url"`
	}, 1) // 向 result.Data 切片中添加一个元素

	// 发送获取任务状态的请求
	for {
		req, err = http.NewRequest("GET", fmt.Sprintf("https://paint.api.wombo.ai/api/v2/tasks/%s", taskResp.ID), nil)
		if err != nil {
			fmt.Println("NewRequest error:", err)
			return nil, err
		}
		req.Header.Set("Authorization", "bearer A")
		resp, err = httpClient.Do(req)
		if err != nil {
			fmt.Println("Request error:", err)
			return nil, err
		}
		defer resp.Body.Close()

		// 解析响应结果
		var taskStateResp TaskStateResponse
		err = json.NewDecoder(resp.Body).Decode(&taskStateResp)
		if err != nil {
			fmt.Println("JSON decode error:", err)
			return nil, err
		}

		// 判断任务状态是否已完成
		if taskStateResp.State == "completed" || taskStateResp.State == "failed" {
			result.Data[0].URL = taskStateResp.Result.Final
			mylog.Logger.Infof("get image url: %s", taskStateResp.Result.Final)
			break
		}

		// 休眠一段时间后再次获取任务状态
		time.Sleep(2 * time.Second)
	}
	if result.Data[0].URL != "" {
		toBase64, err := imageToBase64(result.Data[0].URL)
		if err != nil {
			mylog.Logger.Errorf(" convert to base64 error")
			result.Data[0].URL = ""
		} else {
			result.Data[0].URL = toBase64
		}
	}
	return result, nil

}

func imageToBase64(imgUrl string) (string, error) {

	resp, err := httpClient.Get(imgUrl)
	if err != nil {
		mylog.Logger.Error("Error while downloading image:", err)
		return "", err
	}
	defer resp.Body.Close()

	imageData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mylog.Logger.Error("Error while reading image data:", err)
		return "", err
	}

	base64String := base64.StdEncoding.EncodeToString(imageData)
	return base64String, nil
}

func GenerateImage(prompt string) (*model.GenerationResponse, error) {
	url := "https://api.openai.com/v1/images/generations"
	appkey, err := utils.GetAppKey()
	if err != nil {
		return nil, err
	}
	defer utils.UnlockValue(appkey)

	payload := model.GenerationRequest{
		Prompt: prompt,
		N:      1,
		Size:   "512x512",
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	auth := fmt.Sprintf("Bearer %s", appkey)
	req.Header.Set("Authorization", auth)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := &model.GenerationResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
