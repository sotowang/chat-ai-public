package pay_server

import (
	"fmt"
	"github.com/smartwalle/alipay/v3"
	"testing"
)

func TestPay(t *testing.T) {
	appID := "2016091800539494"
	privateKey := "MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCwaDGlw85S7FLVH8coN3/gPVdalYazVOCf9ITAwXV1jc+cli6SP4eNdIquwGabZWnm7xll88EKBrMPH5ntb437eWEZ42nlfSy8HKN7+HxkHmya5TDAONTQLB97OWsPoLFCB5O0wnOB/Y7D39xUBCHAzcEl8JbKTstt1bnuN/f8yMlHilXvmSp8v9wNCWaowU599BSa7w3YDSRQBwiwB27kX7WeZn1bvQgyFBfHZgRHAaiqCd3TRu/o0jtD4jpCrloHcdfKklmwKUOYVxMpMtCGQucz7oTah22KBgSIFTmLDYYdR4LvVFa8RrEOxfGXrZuCACQ5ZX4XhSPAYGePj2TvAgMBAAECggEBAISiYqnK1gd9Di6WDQzM5cW4+UPvZu7Q428AzwuKqYDwVefa9zsYrxeJR3hqyvsNvhlXLpZ8NmkObnTxgq+GD4/DTGHFnephOlBZjAX657PzOB8iMYxlboTyl9zzZ+7BGGWYAtOG3IghF8VoPGWjEanASff3s7h6k/OKHZjh1f/lyEQnn4PHhoCpngYIlLjDWogrC0/sqPrD7d/PFu9OVVjpcoZnuhtHDsgAeWN51PB2o5cfbE+gGSwY4FhTbyR69NGN9IeARDlOETSa4gnBstg7/JE0ckV8iV/HueftnmFd16ajw150wPWuCvt9pCe0Maj1T37FAn3Pre2YBB0fa6ECgYEA924qQkZ+d6l2IjWj9ULc2fmVjzYNFbBLqX0b6rdEqh/7CG7MbiIsTIH0h7YDXUimrk93qIdqLadpjTlnC0h9V6hUfljU6pZVIO+XvBXzHrszcewhbpdiowiFUHuqaH6rMACJ518yhFTsRqmV7aETrhPjgRDa6g9mBb01FLAo45ECgYEAtoRNZBJjMqvNhyhUybouW3JmmOnA4s+HdHoAINpjM8jyCLZ/Tlke9yics2F/CZbMjDF13DYzbFba4EtZ66lt8u79exTqgULEMDHYufRifAmWSyqGfYUVeIGcf/453/1/k1uHf4x0A9xsvxo14gKPIoaPrjAJGqDNFD+57+aegH8CgYAzPnCoGzt0AvfBbASR9hARYNx1tYcON93js1KF0QD6jvcJrxDNumwcSEnhlXOq7TIAJdstXyZBYEu/AOVzc8bp0aX2KOWn1Ay7boOpY45fjfvAm1vtMJMwGsKpgYMwcxN3NJVbAt9OgtwQYmz3swWFZv8WKux7z0ac56vHphhB0QKBgAhoYdRJUI6GAYrHXdiJHheSVo2Wvw7ztm60LAtXZBh/mj6ygXzPeDC0iztsM1jyvGt838wMJyRHf/+zGOpVPL5jKgQge4kG1VjPAwfV7S9/lY/S3q0rk0ig19/Bi1L5L4ZPQherFbET12KaR38o1QUnI7lHdzPl0myrXtphSk3rAoGBAKKdaipI4kVYtOf8ot+Xd1GZxwsPF53N5dbSi1jkQUcvjtdfBtyJPhqs/d+MJ3Gq194wnDq6XCV08TQ84ZJffWt1kH1R8c8fxR1lyEYQcZ+Jvvw0Lf2AGCfRqWd78HkECOfPFip2+kQnfF8wombtbs8x76zWOZI/Wnchkzw5TOY+"
	aliPublicKey := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2IV01WHUKCIQhJ25jrFFQexhu1nhS0o8m5pDUVAbIhSI3H7SpsBLj/8yC46LhWz/6sbap8gAmWmGxugPCO3NHujTfVwyqkCmwcMHhfnZFgY+jjCtgaYTHl+OLrt3mtSDaSYxs3KCjbJfOW9ohmZ/6ukjniXB0iBvoUIx7ttvob3RB0Nf0mp3wuk/uuLTUFBxomjKyvs15LC0/p11CMWb5ss8iESHDYEoZz8llv2jFvp3jmUzTyySKBEeExvwe+8THTKbThJmPYYaN1Fx+D2UHbiumUEjcVdfPiYEfphFoOQc7bv6cs9RuGjDPM78WHL/sCU0/+RS0pckpN+ECZTPmQIDAQAB"
	isProduction := false
	var client, err = alipay.New(appID, privateKey, isProduction)
	if err != nil {
		panic(err)
	}
	err = client.LoadAliPayPublicKey(aliPublicKey)
	if err != nil {
		panic(err)
	}

	//client.SetEncryptKey("key")
	var param = alipay.TradePreCreate{
		Trade: alipay.Trade{
			Subject:        "识鱼",                   //订单标题
			OutTradeNo:     "eklsfhlkefh",          //商户订单号
			TotalAmount:    "4.98",                 // 订单总金额，单位为元，
			ProductCode:    "FACE_TO_FACE_PAYMENT", //销售产品码，与支付宝签约的产品码名称
			TimeoutExpress: "30m",
		},
	}
	result, err := client.TradePreCreate(param)
	if err != nil {
		panic(err)

	}
	fmt.Printf("%+v", result)
}
