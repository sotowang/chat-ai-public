package model

import "time"

// 定义pdf_record表的结构体
type PDFRecord struct {
	ID        uint `gorm:"primary_key"`
	UserID    int
	DocID     string
	DocType   string
	UploadAt  time.Time
	QaCount   int
	CreatedAt time.Time
	UpdatedAt time.Time
	Filename  string
	Status    int
}

type DocsVO struct {
	DocID    string
	Filename string
}

func (PDFRecord) TableName() string {
	return "pdf_record"
}

type UploadResponse struct {
	Code     int     `json:"code"`
	Message  string  `json:"message"`
	Time     float64 `json:"time"`
	Filename string  `json:"filename"`
	DocId    string  `json:"doc_id"`
}

type AskResponse struct {
	DocID string `json:"doc_id"`
	Msg   string `json:"msg"`
}

type Response struct {
	Answer string `json:"answer"`
}
