package model

import "time"

type QARecord struct {
	ID        int       `json:"id"`
	UserID    uint      `json:"user_id"`
	QADate    string    `json:"qa_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Count     int       `json:"count"`
	Type      int       `json:"type"`
}

func (QARecord) TableName() string {
	return "qa_record"
}
