package repo

import (
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"errors"
	"gorm.io/gorm"
	"time"
)

func CheckVerificationCode(userID uint, code string) (bool, error) {
	vc := model.VerificationCode{}
	if err := MyDB.Where("user_id = ? AND code = ?", userID, code).First(&vc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// 验证码是否过期
	if vc.ExpireTime.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}

func QueryVerifyTimesToday(userId uint) int64 {
	today := time.Now().Format("2006-01-02")
	var count int64
	MyDB.Model(&model.VerificationCode{}).
		Where("user_id = ? AND DATE(created_at) = ?", userId, today).
		Count(&count)
	return count
}

func Insert(userId uint, code string) error {
	// 插入验证码
	vc := model.VerificationCode{
		UserID:     userId,
		Code:       code,
		ExpireTime: time.Now().Add(60 * time.Minute),
	}
	if err := MyDB.Create(&vc).Error; err != nil {
		mylog.Logger.Errorf("Failed to insert verification code!")
		return err
	}
	return nil
}
