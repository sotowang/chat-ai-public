package repo

import (
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"fmt"
	"gorm.io/gorm"
	"time"
)

func GetOrdersByUserID(userID uint64, limit int) ([]*model.Order, error) {
	orders := make([]*model.Order, 0)
	result := MyDB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&orders)
	if result.Error != nil {
		return nil, result.Error
	}
	return orders, nil
}

// CreateOrder is a function that creates an order in the database using gorm
func CreateOrder(order *model.Order) error {

	// Migrate the schema
	err := MyDB.AutoMigrate(&model.Order{})
	if err != nil {
		return err
	}

	// Create the order in the orders table
	err = MyDB.Create(order).Error
	if err != nil {
		return err
	}

	fmt.Println("Order created:", order)

	return nil
}

func UpdateOrderStatus(orderNo string, newStatus int) error {
	// Start a transaction for atomicity
	tx := MyDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update the order status
	if err := tx.Model(&model.Order{}).
		Where("order_no = ?", orderNo).
		Update("status", newStatus).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func UpdateOrderAndAddUserTime(orderNo string, newStatus int) (*model.Order, error) {
	var order = model.Order{}
	var user = model.User{}

	tx := MyDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the order from the database
	result := tx.Where("order_no = ?", orderNo).Set("gorm:query_option", "FOR UPDATE").First(&order)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, gorm.ErrRecordNotFound
	}
	if order.Status == newStatus {
		tx.Rollback()
		return &order, _const.ORDER_STATUS_ALREADY_UPDATE_ERROT
	}

	// Update the order's status
	order.Status = newStatus
	result = tx.Save(&order)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	// Update the user's VIP expire date
	result = tx.Where("id = ?", order.UserID).First(&user)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	var product model.Product
	res := QueryProductById(order.ProductID, &product)
	if res.Error != nil {
		tx.Rollback()
		return nil, res.Error
	}
	now := time.Now().Unix()
	if user.VipExpireDate > now {
		user.VipExpireDate += int64(product.ProductValue)
	} else {
		user.VipExpireDate = now + int64(product.ProductValue)
	}
	user.VipStatus = _const.USER_VIP_STATUS
	result = tx.Save(&user)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	mylog.Logger.Infof("userId: %d, 更新订单: %s 状态为完成, 产品值: %d ", order.UserID, orderNo, product.ProductValue)
	return &order, nil
}
