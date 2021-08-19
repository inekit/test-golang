package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Деплоим xml-файл для теста
	router.GET("/getXML", func(c *gin.Context) {
		c.File("./static/feed_example.xml")
	})

	// Публикуем API для получения структуры, записи в БД и получения из нее данных на порту 8080
	router.POST("/api/parse", server.parseByLink("struct"))
	router.POST("/api/parse/save", parseByLink("db"))
	router.POST("/api/get", getByStruct)
	router.Run("localhost:8080")

}
