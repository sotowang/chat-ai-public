package repo

import (
	"chat-ai/chat-server/model"
	"errors"
	"gorm.io/gorm"
	"time"
)

// 查询某用户今日上传文件的数量及问答次数
func QueryTodayPdfRecord(userID uint) (uploadCount int, err error) {
	// 获取今天的日期
	today := time.Now().Format("2006-01-02")
	// 查询今天该用户上传的文件数量
	var count int64
	err = MyDB.Model(&model.PDFRecord{}).Where("user_id = ? and upload_at >= ?", userID, today).Count(&count).Error
	if err != nil {
		return 0, err
	}
	uploadCount = int(count)

	return uploadCount, nil
}

func QueryQACntByDocId(userID uint, docId string) (int, error) {
	// 查询某用户某docId的问答次数
	var count int
	if err := MyDB.Table("pdf_record").Where("user_id = ? AND doc_id = ?", userID, docId).Select("qa_count").Scan(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func QueryPdfRecordByStatus(userId uint, status, limit int) ([]*model.PDFRecord, error) {
	var records []*model.PDFRecord
	result := MyDB.Where("user_id = ? AND status = ?", userId, status).Limit(limit).Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

func IncreasePdfQACount(userId int, docId string) error {
	var record model.PDFRecord
	if err := MyDB.Where("user_id = ? AND doc_id = ?", userId, docId).First(&record).Error; err != nil {
		return err
	}
	if err := MyDB.Model(&record).Update("qa_count", record.QaCount+1).Error; err != nil {
		return err
	}
	return nil
}

func SetPdfRecordStatusToDeleted(userID uint, docID string) error {
	return MyDB.Table("pdf_record").Where("user_id = ? AND doc_id = ?", userID, docID).
		Update("status", 1).Error
}

func SaveUploadPdfRecord(userId int, docId, filename string, docType string) error {
	// 先查询记录是否存在
	var record model.PDFRecord
	if err := MyDB.Where("doc_id = ?", docId).First(&record).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// 查询出错，直接返回错误
			return err
		}
		// 记录不存在，创建一条新记录
		record = model.PDFRecord{
			UserID:   userId,
			DocID:    docId,
			DocType:  docType,
			UploadAt: time.Now(),
			QaCount:  0,
			Filename: filename,
			Status:   0,
		}
	} else {
		// 记录已存在，更新记录
		record.UserID = userId
		record.DocType = docType
		record.UploadAt = time.Now()
		record.QaCount = 0
		record.Filename = filename
		record.Status = 0
	}
	// 保存记录
	err := MyDB.Save(&record).Error
	return err
}
