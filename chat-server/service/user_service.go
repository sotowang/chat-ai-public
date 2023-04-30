package service

import (
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/repo"
	"chat-ai/chat-server/utils"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

func LoginByPhoneOrEmail(phone, code string) (*model.UserVO, error) {
	var user = &model.User{}
	result := repo.QueryUserByPhoneOrEmail(phone, user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果找不到用户，自动注册一个新用户
			_, err := Register(phone, _const.RANDOM_PASSWORD, config.GlobalConf.Server.InitExpireTime)
			if err != nil {
				mylog.Logger.Errorf("failed to register new user:%s  err: %v", phone, err)
				return nil, fmt.Errorf("failed to register new user: %w", err)
			}
			// 再次尝试查找用户
			result = repo.QueryUserByPhoneOrEmail(phone, user)
			if result.Error != nil {
				return nil, fmt.Errorf("failed to find user after registration: %w", result.Error)
			}
		} else {
			return nil, fmt.Errorf("failed to find user: %w", result.Error)
		}
	}
	if user.Status == _const.USER_BANNED_STATUS {
		return nil, _const.USER_BANNED_ERROR
	}
	if b, err := repo.CheckVerificationCode(user.ID, code); err != nil || !b {
		return nil, fmt.Errorf("invalid verification code")
	}
	accessToken, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		mylog.Logger.Errorf("failed to generate JWT accessToken: %s,err:%v", accessToken, err)
		return nil, fmt.Errorf("failed to generate JWT accessToken: %w", err)
	}
	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		mylog.Logger.Errorf("failed to generate JWT refresh token: %s,err:%v", refreshToken, err)
		return nil, fmt.Errorf("failed to generate JWT refresh token: %w", err)
	}
	var userVo = &model.UserVO{
		Email:         phone,
		VipStatus:     user.VipStatus,
		VipExpireDate: transferTime(user.VipExpireDate),
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
	}
	mylog.Logger.Infof("user:%s login success", phone)
	return userVo, nil
}

func Login(email, password string) (*model.UserVO, error) {
	// find a user by email
	var user = &model.User{}
	result := repo.QueryUserByPhoneOrEmail(email, user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// If user not found, automatically register a new user
			_, err := Register(email, password, config.GlobalConf.Server.InitExpireTime)
			if err != nil {
				mylog.Logger.Errorf("failed to register new user:%s  err: %v", email, err)
				return nil, fmt.Errorf("failed to register new user: %w", err)
			}
			// Try to find the user again
			result = repo.QueryUserByPhoneOrEmail(email, user)
			if result.Error != nil {
				return nil, fmt.Errorf("failed to find user after registration: %w", result.Error)
			}
		} else {
			return nil, fmt.Errorf("failed to find user: %w", result.Error)
		}
	}

	// verify the password
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		return nil, fmt.Errorf("password error")
	}

	// generate a JWT accessToken
	accessToken, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		mylog.Logger.Errorf("failed to generate JWT accessToken: %s,err:%v", accessToken, err)
		return nil, fmt.Errorf("failed to generate JWT accessToken: %w", err)
	}
	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		mylog.Logger.Errorf("failed to generate JWT refresh token: %s,err:%v", refreshToken, err)
		return nil, fmt.Errorf("failed to generate JWT refresh token: %w", err)
	}
	user.VipStatus = _const.USER_NORMAL_STATUS
	if user.VipExpireDate > time.Now().Unix() {
		user.VipStatus = _const.USER_VIP_STATUS
	}
	var userVo = &model.UserVO{
		Email:         email,
		VipStatus:     user.VipStatus,
		VipExpireDate: transferTime(int64(user.VipExpireDate)),
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
	}
	mylog.Logger.Infof("user:%s login success", email)
	return userVo, nil
}

func Register(email, password string, expireDate int) (bool, error) {
	// check if the user already exists
	var existingUser = &model.User{}
	result := repo.QueryUserByPhoneOrEmail(email, existingUser)
	if result.Error == nil {
		mylog.Logger.Errorf("user already exists: %s", email)
		return false, fmt.Errorf("user already exists")
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("failed to check user existence: %w", result.Error)
	}

	// hash the password
	cost := 12 // choose an appropriate cost value
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return false, fmt.Errorf("failed to hash password: %w", err)
	}

	// create a new user
	expire := int64(expireDate) + time.Now().Unix()
	newUser := &model.User{Email: email, Password: hashedPassword, VipExpireDate: expire}
	result = repo.InsertUser(newUser)
	if result.Error != nil {
		return false, fmt.Errorf("failed to create user: %w", result.Error)
	}
	mylog.Logger.Infof("user:%s register success, init expire time:%d", email, expireDate)
	return true, nil
}

func UpdateUserStatus(userId uint64, reChargeTime int) error {
	return repo.UpdateVIPExpireDate(userId, reChargeTime)
}

func GetUserInfo(userId int) (*model.UserVO, error) {
	var user = &model.User{}
	result := repo.QueryById(userId, user)
	if result.Error != nil {
		return nil, result.Error
	}
	if user.VipStatus == _const.USER_NORMAL_STATUS && int64(user.VipExpireDate) > time.Now().Unix() {
		user.VipStatus = _const.USER_VIP_STATUS
	}
	var userVo = &model.UserVO{
		Email:         user.Email,
		VipStatus:     user.VipStatus,
		VipExpireDate: transferTime(int64(user.VipExpireDate)),
	}
	return userVo, nil
}

func transferTime(unixTimestamp int64) string {
	timezone := time.FixedZone("CST", 8*3600) // 设置时区为东八区
	t := time.Unix(unixTimestamp, 0).In(timezone)
	return t.Format("2006-01-02 15:04:05") // 时间格式化
}
