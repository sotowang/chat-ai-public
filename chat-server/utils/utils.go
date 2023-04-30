package utils

import (
	"chat-ai/chat-server/config"
	mylog "chat-ai/chat-server/log"
	"fmt"
	"math/rand"
	"time"
)

func GetAppKey() (string, error) {
	apiKeys := config.GlobalConf.OpenAi.AppKeys
	vipAppkeys := config.GlobalConf.OpenAi.VipAppKeys
	shuffle(apiKeys)
	if len(vipAppkeys) > 0 {
		shuffle(vipAppkeys)
		mylog.Logger.Infof("当前获取到vip ak: [%s]", vipAppkeys[0])
		return vipAppkeys[0], nil
	}
	sk, err := getUnlockedValue(apiKeys, AppKeyLock)
	if err != nil {
		mylog.Logger.Errorf("获取ak失败: %v", err)
	} else {
		mylog.Logger.Infof("当前获取到ak: [%s]", sk)
	}
	return sk, err
}

func GetDocAppKey() (string, error) {
	apiKeys := config.GlobalConf.PDF.AppKeys
	vipAppkeys := config.GlobalConf.PDF.VipAppKeys
	shuffle(apiKeys)
	if len(vipAppkeys) > 0 {
		shuffle(vipAppkeys)
		mylog.Logger.Infof("当前获取到vip ak: [%s]", vipAppkeys[0])
		return vipAppkeys[0], nil
	}
	sk, err := getUnlockedValue(apiKeys, AppKeyLock)
	if err != nil {
		mylog.Logger.Errorf("获取ak失败: %v", err)
	} else {
		mylog.Logger.Infof("当前获取到ak: [%s]", sk)
	}
	return sk, err
}

func shuffle(strs []string) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(strs), func(i, j int) {
		strs[i], strs[j] = strs[j], strs[i]
	})
}

func getUnlockedValue(values []string, lock *Lock) (string, error) {
	retries := 3
	for retries > 0 {
		for _, value := range values {
			if lock.Lock(value) {
				return value, nil
			}
		}
		retries--
		time.Sleep(500 * time.Millisecond)
	}
	return "", fmt.Errorf("failed to get an unlocked value")
}

func UnlockValue(value string) {
	AppKeyLock.Unlock(value)
}
