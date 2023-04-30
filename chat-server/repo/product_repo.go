package repo

import (
	"chat-ai/chat-server/model"
	"gorm.io/gorm"
)

/**
根据id查商品信息
*/
func QueryProductById(id uint64, product *model.Product) *gorm.DB {
	return MyDB.Where("id = ? ", id).First(product)
}

/**
查询商品列表
*/
func QueryAllProduct(products *[]model.Product) *gorm.DB {
	// 获取所有产品数据
	return MyDB.Where("status = ?", 0).
		Order("product_id asc").
		Find(&products)
}
