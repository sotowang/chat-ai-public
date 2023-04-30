package model

import "time"

type User struct {
	ID            uint   `gorm:"primaryKey" json:"ID,omitempty"`
	Email         string `gorm:"unique"`
	Password      []byte `json:"password,omitempty"`
	VipStatus     int
	VipExpireDate int64
	Status        int
}

type UserVO struct {
	Email         string
	VipStatus     int
	VipExpireDate string
	AccessToken   string
	RefreshToken  string
}
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TranslationResult struct {
	TranslatedText string
	ElapsedTime    int64
}

type ChatResult struct {
	Code        int
	Article     string
	ElapsedTime int64
}

type PluginResult struct {
	Answer string
}

type RequestMessage struct {
	Ask    string `json:"ask"`
	Answer string `json:"answer"`
}

type RequestMessages struct {
	Role     string           `json:"role"`
	Messages []RequestMessage `json:"messages"`
}

type GenerationRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Object  string   `json:"object"`
	Choices []Choice `json:"choices"`
}

type GenerationResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL string `json:"url"`
	} `json:"data"`
}

type Choice struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`

	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

type CompletionRequest struct {
	Model            string  `json:"model"`
	Prompt           string  `json:"prompt"`
	Temperature      float64 `json:"temperature"`
	MaxTokens        int     `json:"max_tokens"`
	TopP             float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
}

type ModerationResponse struct {
	Id      string `json:"id"`
	Model   string `json:"model"`
	Results []struct {
		Flagged        bool               `json:"flagged"`
		Categories     map[string]bool    `json:"categories"`
		CategoryScores map[string]float64 `json:"category_scores"`
	} `json:"results"`
}

type RolePrompt struct {
	ID          uint `gorm:"primaryKey"`
	Role        string
	Prompt      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Description string
}

type RolePromptVo struct {
	Role        string
	Description string
}
