package repo

import (
	"chat-ai/chat-server/model"
	"gorm.io/gorm"
	"time"
)

func QueryTodayQARecord(userID uint, qaType int) (count int, err error) {
	today := time.Now().Format("2006-01-02")
	err = MyDB.Model(&model.QARecord{}).Where("user_id = ? and qa_date = ? and type = ?", userID, today, qaType).
		Select("count").First(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func IncreaseAskQACount(userId uint, qaType int) error {
	var record model.QARecord
	today := time.Now().Format("2006-01-02")
	// 以事务的方式进行原子性更新
	return MyDB.Transaction(func(tx *gorm.DB) error {
		// 查找今天的问答记录
		err := tx.Model(&model.QARecord{}).Where("user_id = ? and qa_date = ? and type = ? ", userId, today, qaType).First(&record).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// 如果记录不存在，创建一条新的记录
		if err == gorm.ErrRecordNotFound {
			record.UserID = userId
			record.QADate = today
			record.Count = 1
			record.Type = qaType
			return tx.Create(&record).Error
		}

		// 记录已存在，则将count加1
		record.Count++
		return tx.Save(&record).Error
	})
}
