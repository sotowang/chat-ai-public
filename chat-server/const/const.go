package _const

import "fmt"

const (
	VIOLATE_CONTENT          = "检测到文本包含违禁内容,请文明用语"
	REFRESH_TOKEN_SECRET_KEY = "skdfjksdljsldjf"
	ACCESS_TOKEN_SECRET_KEY  = "sjkdhfksdhkjhdskfh"
	ALIPAY_PRODUCT_CODE      = "FACE_TO_FACE_PAYMENT"
	RANDOM_PASSWORD          = "sifjsdfjksfslsdhfkdyds"
	PLUGIN_SALT              = "sdkjfksdjfkjsg4"
	ORDER_NOPAY_STATUS       = 0
	ORDER_SUCCESS_STATUS     = 1
	ORDER_FAILED_STATUS      = 2

	QA_TYPE_ASK   = 1
	QA_TYPE_IMAGE = 2

	USER_VIP_STATUS    = 1 //会员
	USER_NORMAL_STATUS = 0 //会员
	USER_BANNED_STATUS = 1

	REQUEST_FREQUENTLY = 419

	TOKEN_EXPIRE_ERROR = "Token is expired"
	STYLE_ANIME        = 46
	REALISTIC          = 78
	WATER_COLOR        = 91
)

var ORDER_STATUS_ALREADY_UPDATE_ERROT = fmt.Errorf("order status already updated")
var VERIFY_CODE_LIMIT = fmt.Errorf("veryfy code limit error")
var USER_BANNED_ERROR = fmt.Errorf("该用户已被封禁")
var INVALID_URL = fmt.Errorf("invalid url")
