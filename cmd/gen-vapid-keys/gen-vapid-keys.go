package main

import (
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	vapidPrivateKey, vapidPublicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		fmt.Printf("生成VAPID密钥失败: %v\n", err)
		return
	}
	fmt.Printf("私钥放后端.env\nVAPID_PRIVATE_KEY=%s\n", vapidPrivateKey)
	fmt.Printf("公钥放前端.env\nVAPID_PUBLIC_KEY=%s\n", vapidPublicKey)
}
