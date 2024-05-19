package main

import (
	"ham3/middlewares"
	"ham3/routers"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func main() {
	router := gin.Default()

	// middlewareの設定
	router.Use(middlewares.TracingMiddleware())

	// 静的ファイルの設定
	router.Static("/static", "./static")

	// テンプレートファイルの読み込み
	router.LoadHTMLGlob("templates/*")

	// ルーティングの設定
	routers.SetupRouter(router)

	// サーバーの起動
	router.Run(":8081")
}
