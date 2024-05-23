package main

import (
	"ham3/routers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// middlewareの設定

	// 静的ファイルの設定
	router.Static("/static", "./static")

	// テンプレートファイルの読み込み
	router.LoadHTMLGlob("templates/*")

	// ルーティングの設定
	routers.SetupRouter(router)

	// サーバーの起動
	router.Run(":8081")
}
