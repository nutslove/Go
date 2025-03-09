package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Middleware1() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("M1 開始")
		// c.Next()
		// fmt.Println("M1 終了") // 後処理1
	}
}

func Middleware2() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("M2 開始")
		// c.Next()
		// fmt.Println("M2 終了") // 後処理2
	}
}

func main() {
	r := gin.New()
	r.Use(Middleware1(), Middleware2())

	r.GET("/", func(c *gin.Context) {
		fmt.Println("ルートハンドラ実行")
	})

	r.Run(":8080")
}
