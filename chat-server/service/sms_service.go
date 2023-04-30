package service

import (
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/repo"
	"errors"
	"fmt"
	"github.com/alibabacloud-go/darabonba-openapi/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dm20151123 "github.com/alibabacloud-go/dm-20151123/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"gorm.io/gorm"
	"math/rand"
	"strings"
	"time"
)

var SmsClient *dysmsapi20170525.Client
var EmailClient *dm20151123.Client

var lastMsgTime = make(map[string]int64, 0)

func InitSmsClient(accessKeyId, accessKeySecret string) error {
	c1 := &client.Config{
		AccessKeyId:     &accessKeyId,
		AccessKeySecret: &accessKeySecret,
		Endpoint:        tea.String("dysmsapi.aliyuncs.com")}
	newClient1, err1 := dysmsapi20170525.NewClient(c1)
	if err1 != nil {
		return err1
	}
	SmsClient = newClient1

	config := &openapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: &accessKeyId,
		// 必填，您的 AccessKey Secret
		AccessKeySecret: &accessKeySecret,
	}
	// 访问的域名
	config.Endpoint = tea.String("dm.aliyuncs.com")
	_result := &dm20151123.Client{}
	var _err error
	_result, _err = dm20151123.NewClient(config)
	if _err != nil {
		return _err
	}
	EmailClient = _result
	return nil
}

func SendEmailMsg(email string) (int64, error) {
	//查询userId
	var existingUser = &model.User{}
	result := repo.QueryUserByPhoneOrEmail(email, existingUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果找不到用户，自动注册一个新用户
			_, err := Register(email, _const.RANDOM_PASSWORD, config.GlobalConf.Server.InitExpireTime)
			if err != nil {
				mylog.Logger.Errorf("failed to register new user:%s  err: %v", email, err)
				return 0, fmt.Errorf("failed to register new user: %w", err)
			}
			// 再次尝试查找用户
			result = repo.QueryUserByPhoneOrEmail(email, existingUser)
			if result.Error != nil {
				return 0, fmt.Errorf("failed to find user after registration: %w", result.Error)
			}
		} else {
			return 0, fmt.Errorf("failed to find user: %w", result.Error)
		}
	}
	if existingUser.Status == _const.USER_BANNED_STATUS {
		return 0, _const.USER_BANNED_ERROR
	}
	remainTime := getLastMsgTime(email)
	if remainTime > 0 {
		mylog.Logger.Errorf("%s 验证码获取剩余时间:%d s", email, remainTime)
		return remainTime, nil
	}
	codeStr := GenerateSmsCode(6)
	code := fmt.Sprintf("识鱼验证码:\n %s \n-----60分钟内有效，请勿泄漏给他人-----", codeStr)

	singleSendMailRequest := &dm20151123.SingleSendMailRequest{
		AccountName:    tea.String(config.GlobalConf.SMS.EmailFrom),
		AddressType:    tea.Int32(1),
		Subject:        tea.String("识鱼 Auth Code"),
		ToAddress:      tea.String(email),
		ReplyToAddress: tea.Bool(false),
		TextBody:       tea.String(code),
	}
	resp, err := EmailClient.SingleSendMail(singleSendMailRequest)
	if err != nil {
		mylog.Logger.Errorf("发送邮箱验证码失败,err: %v", err)
		return 0, err
	}
	if *resp.StatusCode == 200 {
		err = repo.Insert(existingUser.ID, codeStr)
		if err != nil {
			mylog.Logger.Errorf("保存验证码到数据库失败：mail:%s, code: %s,  err: %v", email, codeStr, err)
			return 0, err
		}
		mylog.Logger.Infof("发送邮箱验证码给：%s 成功： code: %v", email, codeStr)
		saveMsgTime(email)
		return 0, nil
	} else {
		mylog.Logger.Errorf("发送邮箱验证码失败: %v ", *resp)
		return 0, nil
	}
}

func saveMsgTime(email string) {
	lastMsgTime[email] = time.Now().Unix()
}

func getLastMsgTime(email string) int64 {
	lastTime, ok := lastMsgTime[email]
	if !ok {
		return 0
	}
	now := time.Now().Unix()
	if lastTime > 0 && now-lastTime < int64(config.GlobalConf.SMS.Frequent) {
		return int64(config.GlobalConf.SMS.Frequent) - (now - lastTime)
	}
	return 0
}

func SendMsg(phone, sign, templateCode string) error {
	//查询userId
	var existingUser = &model.User{}
	result := repo.QueryUserByPhoneOrEmail(phone, existingUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果找不到用户，自动注册一个新用户
			_, err := Register(phone, _const.RANDOM_PASSWORD, config.GlobalConf.Server.InitExpireTime)
			if err != nil {
				mylog.Logger.Errorf("failed to register new user:%s  err: %v", phone, err)
				return fmt.Errorf("failed to register new user: %w", err)
			}
			// 再次尝试查找用户
			result = repo.QueryUserByPhoneOrEmail(phone, existingUser)
			if result.Error != nil {
				return fmt.Errorf("failed to find user after registration: %w", result.Error)
			}
		} else {
			return fmt.Errorf("failed to find user: %w", result.Error)
		}
	}
	if existingUser.Status == _const.USER_BANNED_STATUS {
		return _const.USER_BANNED_ERROR
	}
	codeStr := GenerateSmsCode(6)
	code := "{\"code\":" + codeStr + "}"
	//判断今日发送短信次数
	cnt := repo.QueryVerifyTimesToday(existingUser.ID)
	if cnt >= 3 {
		mylog.Logger.Errorf("用户: %s 今日验证码已达上线", phone)
		return _const.VERIFY_CODE_LIMIT
	}
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(sign),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(code),
	}
	res, err := SmsClient.SendSms(sendSmsRequest)
	if err != nil {
		mylog.Logger.Errorf("阿里云发送短信给：%s 失败： err: %v", phone, err)
		return err
	}
	if *res.Body.Code == "OK" {
		err = repo.Insert(existingUser.ID, codeStr)
		if err != nil {
			mylog.Logger.Errorf("保存短信到数据库失败：phone:%s, code: %s,  err: %v", phone, codeStr, err)
			return err
		}
		mylog.Logger.Infof("阿里云发送短信给：%s 成功： code: %v", phone, codeStr)
		return nil
	} else {
		mylog.Logger.Errorf("阿里云发送短信失败: %s ", *res.Body.Message)
		return nil
	}
}

// GenerateSmsCode 生成验证码;length代表验证码的长度
func GenerateSmsCode(length int) string {
	numberic := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	rand.Seed(time.Now().Unix())
	var sb strings.Builder
	for i := 0; i < length; i++ {
		fmt.Fprintf(&sb, "%d", numberic[rand.Intn(len(numberic))])
	}
	return sb.String()
}
