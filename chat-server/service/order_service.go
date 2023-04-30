package service

import (
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	pay_server "chat-ai/chat-server/pay-server"
	"chat-ai/chat-server/repo"
	"encoding/base64"
	"fmt"
	"github.com/skip2/go-qrcode"
	"github.com/smartwalle/alipay/v3"
	"log"
	"math/rand"
	"time"
)

func GetOrderByUserId(userId uint) ([]*model.OrderVO, error) {
	orders, err := repo.GetOrdersByUserID(uint64(userId), 5)
	if err != nil {
		mylog.Logger.Errorf("get user:%d order err:%v", userId, err)
		return nil, err
	}
	orderVOs := make([]*model.OrderVO, 0, len(orders))
	for _, order := range orders {
		qrUrlCode := ""
		if order.Status == _const.ORDER_NOPAY_STATUS {
			if checkOrderExpire(order) {
				mylog.Logger.Infof("订单id: %s 已失效", order.OrderNo)
				continue
			}
			status, err := pay_server.QueryOrderStatus(order.OrderNo)
			if err != nil {
				mylog.Logger.Errorf("获取订单: %s 状态失败：%v", order.OrderNo, status)
			} else {
				if status == alipay.TradeStatusClosed {
					order.Status = _const.ORDER_FAILED_STATUS
					//更新订单状态为关闭
					UpdateOrderStatus(order.OrderNo, _const.ORDER_FAILED_STATUS)
					mylog.Logger.Infof("更新订单:%s 状态为关闭", order.OrderNo)
				} else if status == alipay.TradeStatusSuccess {
					//更新订单状态为完成
					mylog.Logger.Infof("获取订单信息:%+v", order)
					_, err = repo.UpdateOrderAndAddUserTime(order.OrderNo, _const.ORDER_SUCCESS_STATUS)
					if err != nil {
						mylog.Logger.Errorf("GetOrderByUserId 更新订单:%s 失败", order.OrderNo)
					} else {
						mylog.Logger.Infof("GetOrderByUserId 更新订单: %s 状态为完成", order.OrderNo)
					}
				} else { //如果没查到或订单状态为待支付
					if order.QrUrl != "" {
						qrUrlCode = ImageToBase64(order.QrUrl)
					}
					mylog.Logger.Infof("当前订单: %s 状态为: %s ", order.OrderNo, status)
				}
			}
		}

		orderVO := &model.OrderVO{
			Status:     order.Status,
			QrURL:      order.QrUrl,
			TotalPrice: order.TotalPrice,
			OrderNo:    order.OrderNo,
			CreatedAt:  transferTime(order.CreatedAt.Unix()),
			QR:         qrUrlCode,
		}
		orderVOs = append(orderVOs, orderVO)
	}

	return orderVOs, nil
}

func checkOrderExpire(order *model.Order) bool {
	//1.判断订单创建时间距现在是否大于30min了
	if time.Now().Sub(order.CreatedAt) > 30*time.Minute {
		//2.若是，则更新订单状态为失效，返回true
		UpdateOrderStatus(order.OrderNo, _const.ORDER_FAILED_STATUS)
		return true
	}
	return false
}

func UpdateOrderStatus(orderNo string, newStatus int) error {
	err := repo.UpdateOrderStatus(orderNo, newStatus)
	if err != nil {
		mylog.Logger.Errorf("update order: %s  status: %d  err:%v", orderNo, newStatus, err)
		return err
	}
	return nil
}

func UpdateOrderStatusAndAddUserTime(orderNo string, newStatus int) (*model.Order, error) {
	order, err := repo.UpdateOrderAndAddUserTime(orderNo, newStatus)
	if err != nil {
		mylog.Logger.Errorf("update order: %s  status: %d  err:%v", orderNo, newStatus, err)
		return nil, err
	}
	return order, nil
}

func CreateOrder(userID uint64, productID uint64, quantity int, source string, payType int) (*model.OrderVO, error) {
	// 生成订单号
	orderNo := generateOrderNo()

	// 查询商品信息
	var product = &model.Product{}
	result := repo.QueryProductById(productID, product)
	if result.Error != nil {
		mylog.Logger.Errorf("product id:%d not exist", productID)
		return nil, result.Error
	}

	// 计算订单总价
	totalPrice := product.Price * float64(quantity)
	var order = &model.Order{
		UserID:     userID,
		OrderNo:    orderNo,
		ProductID:  productID,
		TotalPrice: totalPrice,
		Source:     source,
		PayType:    payType,
	}
	//支付码
	amount := fmt.Sprintf("%.2f", order.TotalPrice)
	res, err := pay_server.CreatePay(amount, orderNo, product.ProductName)
	if err != nil {
		mylog.Logger.Errorf("获取支付宝二维码失败,err:%v", err)
		return nil, err
	}
	order.QrUrl = res.Content.QRCode

	err = repo.CreateOrder(order)
	if err != nil {
		mylog.Logger.Errorf("创建订单失败,err:%v", err)
		return nil, err
	}
	var orderVo = &model.OrderVO{
		Status: 200,
		QR:     ImageToBase64(res.Content.QRCode),
		QrURL:  res.Content.QRCode,
	}

	return orderVo, nil
}

func ImageToBase64(text string) string {
	// 要转换的文字
	// 生成二维码图片大小
	size := 256
	// 生成二维码图片字节切片
	png, err := qrcode.Encode(text, qrcode.Medium, size)
	if err != nil {
		log.Fatal(err)
	}
	// 将二维码图片字节切片转成base64编码字符串
	return base64.StdEncoding.EncodeToString(png)
}

func generateOrderNo() string {
	// Get the current time in format YYYYMMDDHHMMSS
	now := time.Now().Format("20060102150405")

	// Generate a random number between 1000 and 9999
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(9000) + 1000

	// Concatenate the time and the random number
	orderNo := fmt.Sprintf("%s%d", now, randNum)

	return orderNo
}
