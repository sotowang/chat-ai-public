package pay_server

import (
	"chat-ai/chat-server/config"
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"fmt"
	"github.com/smartwalle/alipay/v3"
	"net/http"
	"net/url"
)

var alipayClient *alipay.Client

func CreatePayClient() error {
	appID := config.GlobalConf.Pay.AppId
	privateKey := config.GlobalConf.Pay.PrivateKey
	aliPublicKey := config.GlobalConf.Pay.AliPublicKey
	isProduction := config.GlobalConf.Pay.IsProduction
	var client, err = alipay.New(appID, privateKey, isProduction)
	if err != nil {
		mylog.Logger.Errorf("初始化支付client失败,err:[%v]", err)
		return err
	}
	err = client.LoadAliPayPublicKey(aliPublicKey)
	if err != nil {
		mylog.Logger.Errorf("加载支付宝公钥失败,err:[%v]", err)
		return err
	}
	alipayClient = client
	return nil
}

func CreatePay(amount string, orderNo string, subject string) (*alipay.TradePreCreateRsp, error) {
	timeout := config.GlobalConf.Pay.Timeout
	//client.SetEncryptKey("key")
	var param = alipay.TradePreCreate{
		Trade: alipay.Trade{
			Subject:        fmt.Sprintf("识鱼-%s", subject), //订单标题
			OutTradeNo:     orderNo,                       //商户订单号
			TotalAmount:    amount,                        // 订单总金额，单位为元，
			ProductCode:    _const.ALIPAY_PRODUCT_CODE,    //销售产品码，与支付宝签约的产品码名称
			TimeoutExpress: timeout,
			NotifyURL:      config.GlobalConf.Pay.NotifyUrl,
			ReturnURL:      config.GlobalConf.Pay.ReturnUrl,
		},
	}
	result, err := alipayClient.TradePreCreate(param)
	if err != nil || !result.Content.Code.IsSuccess() {
		mylog.Logger.Errorf("创建支付失败:%+v ,err:%v", result, err)
		return nil, err
	}
	return result, nil
}

func VerifySign(data url.Values) (bool, error) {
	return alipayClient.VerifySign(data)
}

func GetNotify(req *http.Request) *alipay.TradeNotification {
	var noti, _ = alipayClient.GetTradeNotification(req)
	return noti
}

func QueryOrderStatus(orderNo string) (alipay.TradeStatus, error) {
	// Create a TradeQuery struct with the necessary parameters
	query := alipay.TradeQuery{}
	query.OutTradeNo = orderNo

	// Call the TradeQuery method to query the order status
	resp, err := alipayClient.TradeQuery(query)
	if err != nil {
		mylog.Logger.Errorf("查询订单状态失败，err:[%v]", err)
		return "", err
	}
	return resp.Content.TradeStatus, nil

}
